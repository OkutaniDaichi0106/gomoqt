package moqt

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/internal/transport"
)

func NewRelayer(bufferSize int) *Relayer {
	relayer := &Relayer{
		servingTrackNames: make(map[SubscribeID]string),
		upstream:          make(map[string]*SentSubscription),
		downstreams:       make(map[string][]*ReceivedSubscription),
		dataQueue:         make(dataQueue, 0, 1<<4),
		ch:                make(chan struct{}, 1),
		BufferSize:        bufferSize,
	}

	// Initialize the relayer
	relayer.run()

	return relayer
}

type Relayer struct {
	// Track Namespace
	servingTrackNames map[SubscribeID]string

	/*
	 * The upstream
	 * Track Name -> Sent Subscription
	 */
	upstream map[string]*SentSubscription
	usMu     sync.RWMutex

	/*
	 * The downstreams
	 * The key is the upstream's subscribe ID
	 * Track Name -> downstreams
	 */
	downstreams map[string][]*ReceivedSubscription
	dsMu        sync.RWMutex

	// The track aliases
	// trackAliases map[string]SubscribeID

	// The data queue
	dataQueue dataQueue
	ch        chan struct{}

	// The buffer size
	BufferSize int
}

func (r *Relayer) addUpstream(trackName string, upstream *SentSubscription) {
	r.usMu.Lock()
	defer r.usMu.Unlock()

	r.servingTrackNames[upstream.SubscribeID()] = trackName

	r.upstream[trackName] = upstream
}

func (r *Relayer) removeUpstream(trackName string) {
	r.usMu.Lock()
	defer r.usMu.Unlock()

	delete(r.upstream, trackName)
}

func (r *Relayer) addDownstream(trackName string, downstream *ReceivedSubscription) {
	r.dsMu.Lock()
	defer r.dsMu.Unlock()

	// Get the downstreams
	downstreams, ok := r.downstreams[trackName]
	if !ok {
		downstreams = make([]*ReceivedSubscription, 0, 1)
	}

	// Append the downstream
	downstreams = append(downstreams, downstream)

	// Update the downstreams
	r.downstreams[trackName] = downstreams
}

func (r *Relayer) removeDownstream(trackName string, downstream *ReceivedSubscription) {
	r.dsMu.Lock()
	defer r.dsMu.Unlock()

	// Get the downstreams
	downstreams, ok := r.downstreams[trackName]
	if !ok {
		return
	}

	// Remove the downstream
	for i, ds := range downstreams {
		if ds == downstream {
			downstreams = append(downstreams[:i], downstreams[i+1:]...)
			break
		}
	}

	// Update the downstreams
	r.downstreams[trackName] = downstreams
}

func (r *Relayer) run() {
	ctx := context.TODO() // TODO: context

	// Listen for data streams
	go r.listenStreamData(ctx)

	// Listen for datagrams
	go r.listenDatagram(ctx)

	// Start the data distribution
	go r.distribute(ctx)
}

func (r *Relayer) listenStreamData(ctx context.Context) {
	for {
		for _, subscription := range r.upstream {
			go func(subscription *SentSubscription) {
				// Receive data from the upstream
				stream, err := subscription.AcceptDataStream(ctx)
				if err != nil {
					slog.Error("failed to receive data from the upstream", slog.String("error", err.Error()))
					return
				}

				go func(stream DataReceiveStream) {
					for {
						buf := make([]byte, r.BufferSize*(1<<10))
						// Read the data from the stream
						n, err := stream.Read(buf)
						if err != nil {
							slog.Error("failed to read data from the stream", slog.String("error", err.Error()))
							return
						}

						// Create a new data object
						data := &streamDataFragment{
							trackPriority: subscription.TrackPriority,
							groupOrder:    subscription.GroupOrder,
							receivedGroup: receivedGroup{
								subscribeID:   stream.SubscribeID(),
								groupSequence: stream.GroupSequence(),
								groupPriority: stream.GroupPriority(),
								receivedAt:    time.Now(),
							},
							streamID: stream.StreamID(),
							payload:  buf[:n],
						}

						// Enqueue the data
						r.dataQueue.Push(data)

						// Notify the data distribution
						r.ch <- struct{}{}
					}
				}(stream)
			}(subscription)
		}
	}
}

func (r *Relayer) listenDatagram(ctx context.Context) {
	for {
		for _, subscription := range r.upstream {
			go func(subscription *SentSubscription) {
				// Receive data from the upstream
				datagram, err := subscription.AcceptDatagram(ctx)
				if err != nil {
					slog.Error("failed to receive data from the upstream", slog.String("error", err.Error()))
					return
				}

				// Create a new data object
				data := &datagramData{
					trackPriority: subscription.TrackPriority,
					groupOrder:    subscription.GroupOrder,
					receivedGroup: receivedGroup{
						subscribeID:   datagram.SubscribeID(),
						groupSequence: datagram.GroupSequence(),
						groupPriority: datagram.GroupPriority(),
						receivedAt:    datagram.ReceivedAt(),
					},
					payload: datagram.Payload(),
				}

				// Enqueue the data
				r.dataQueue.Push(data)

				// Notify the data distribution
				r.ch <- struct{}{}
			}(subscription)
		}
	}
}

func (r *Relayer) distribute(ctx context.Context) {
	allStreams := make(map[transport.StreamID][]DataSendStream)

	// Distribute data
	for {
		if r.dataQueue.Len() > 0 {
			// Dequeue the data
			data := r.dataQueue.Pop().(dataFragment)

			trackName, ok := r.servingTrackNames[data.SubscribeID()]
			if !ok {
				slog.Error("track does not exist", slog.Any("subscribeID", data.SubscribeID()))
				continue
			}

			// Get the downstreams for the subscribe ID
			subscriptions, ok := r.downstreams[trackName]
			if !ok || len(subscriptions) == 0 {
				slog.Error("no downstreams", slog.Any("subscribeID", data.SubscribeID()))
				continue
			}

			// Handle the data
			switch data := data.(type) {
			case *streamDataFragment:
				// Verify if servingStreams with the same ID exist
				servingStreams, ok := allStreams[data.StreamID()]
				if !ok {
					if len(servingStreams) == 0 {
						servingStreams = make([]DataSendStream, 0, len(subscriptions))
					}

					for _, rs := range subscriptions {
						// Open a new data stream
						stream, err := rs.OpenDataStream(data.GroupSequence(), data.GroupPriority())
						if err != nil {
							slog.Error("failed to open a data stream", slog.String("error", err.Error()))
							continue
						}

						// Append the stream
						servingStreams = append(servingStreams, stream)
					}

					// Register the streams
					allStreams[data.StreamID()] = servingStreams
				}

				// Write the data to the stream
				for _, stream := range servingStreams {
					go func(stream DataSendStream) {
						// Write the data to the stream
						_, err := stream.Write(data.Payload())
						if err != nil {
							slog.Error("failed to write data to the stream", slog.String("error", err.Error()))
							return
						}
					}(stream)
				}

			case *datagramData:

				for _, subscription := range subscriptions {
					// Send the data to the downstream
					_, err := subscription.SendDatagram(data.SubscribeID(), data.GroupSequence(), data.GroupPriority(), data.Payload())
					if err != nil {
						slog.Error("failed to write data to the stream", slog.String("error", err.Error()))
						return
					}
				}
			}
		}

		// Wait for the next data
		select {
		case <-ctx.Done():
			return
		case <-r.ch:
			continue
		default:
		}
	}
}
