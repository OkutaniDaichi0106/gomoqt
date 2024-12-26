package moqt

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/internal/transport"
)

type relayer interface {
	relay()
}

var _ relayer = (*Relayer)(nil)

func NewRelayer(upstream *SentSubscription) *Relayer {
	relayer := &Relayer{
		upstream: upstream,
	}

	relayer.init()

	return relayer
}

type Relayer struct {
	// The upstream
	upstream map[SubscribeID]*SentSubscription

	// The downstreams
	downstreams map[SubscribeID][]*ReceivedSubscription
	dsMu        sync.RWMutex

	// The data queue
	dataQueue dataQueue
	ch        chan struct{}

	// The buffer size
	BufferSize int
}

func (r *Relayer) init() {
	r.downstreams = make(map[SubscribeID][]*ReceivedSubscription)
	r.dataQueue = make(dataQueue, 0, 1<<4)
	r.ch = make(chan struct{}, 1)

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
		select {
		case <-ctx.Done():
			return
		default:
			if r.dataQueue.Len() > 0 {
				data := r.dataQueue.Pop().(dataFragment)
				switch data.(type) {
				case *streamDataFragment:
					streamID := data.(*streamDataFragment).StreamID()

					subscriptions, ok := r.downstreams[data.SubscribeID()]
					if !ok || len(subscriptions) == 0 {
						slog.Error("no downstreams", slog.Any("subscribeID", data.SubscribeID()))
						continue
					}

					// Verify if streams with the same ID exist
					streams, ok := allStreams[streamID]
					if !ok {
						if len(streams) == 0 {
							streams = make([]DataSendStream, 0, len(subscriptions))
						}

						for _, rs := range subscriptions {
							// Open a new data stream
							stream, err := rs.OpenDataStream(data.GroupSequence(), data.GroupPriority())
							if err != nil {
								slog.Error("failed to open a data stream", slog.String("error", err.Error()))
								continue
							}

							streams = append(streams, stream)
						}
					}

					// Write the data to the stream
					for _, stream := range streams {
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
					subscriptions, ok := r.downstreams[data.SubscribeID()]
					if !ok || len(subscriptions) == 0 {
						slog.Error("no downstreams", slog.Any("subscribeID", data.SubscribeID()))
						continue
					}

					for _, rs := range subscriptions {
						// Send the data to the downstream
						_, err := rs.SendDatagram(data.SubscribeID(), data.GroupSequence(), data.GroupPriority(), data.Payload())
						if err != nil {
							slog.Error("failed to write data to the stream", slog.String("error", err.Error()))
							return
						}
					}
				}
			}
			select {
			case <-r.ch:
				continue
			}
		}
	}
}

func (r *Relayer) addDownstream(rs *ReceivedSubscription) {
	r.dsMu.Lock()
	defer r.dsMu.Unlock()

	r.downstreams[rs.SubscribeID()] = append(r.downstreams[rs.SubscribeID()], rs)
}

func (r *Relayer) removeDownstream(rs *ReceivedSubscription) {
	r.dsMu.Lock()
	defer r.dsMu.Unlock()

	subscriptions, ok := r.downstreams[rs.SubscribeID()]
	if !ok || len(subscriptions) == 0 {
		return
	}

	for i, subscription := range subscriptions {
		if subscription == rs {
			subscriptions = append(subscriptions[:i], subscriptions[i+1:]...)
			r.downstreams[rs.SubscribeID()] = subscriptions
			break
		}
	}
}
