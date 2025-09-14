package moqt

import (
	"context"
	"sync"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/quic"
)

func newReceiveSubscribeStream(id SubscribeID, stream quic.Stream, config *TrackConfig) *receiveSubscribeStream {
	// Ensure config is not nil
	if config == nil {
		config = &TrackConfig{}
	}

	rss := &receiveSubscribeStream{
		subscribeID: id,
		config:      config,
		stream:      stream,
		updatedCh:   make(chan struct{}, 1),
		ctx:         context.WithValue(stream.Context(), &biStreamTypeCtxKey, message.StreamTypeSubscribe),
	}

	// Listen for updates in a separate goroutine
	go func() {
		var sum message.SubscribeUpdateMessage
		var err error

		for {
			rss.configMu.Lock()
			if rss.ctx.Err() != nil {
				rss.configMu.Unlock()
				break
			}
			rss.configMu.Unlock()

			err = sum.Decode(rss.stream)
			if err != nil {
				break
			}

			rss.configMu.Lock()
			rss.config = &TrackConfig{
				TrackPriority:    sum.TrackPriority,
				MinGroupSequence: sum.MinGroupSequence,
				MaxGroupSequence: sum.MaxGroupSequence,
			}

			select {
			case rss.updatedCh <- struct{}{}:
			default:
			}
			rss.configMu.Unlock()
		}

		// Cleanup after loop ends
		rss.configMu.Lock()
		rss.updatedOnce.Do(func() {
			if rss.updatedCh != nil {
				close(rss.updatedCh)
			}
		})
		rss.configMu.Unlock()

	}()

	return rss
}

type receiveSubscribeStream struct {
	subscribeID SubscribeID

	stream quic.Stream

	acceptOnce sync.Once

	configMu    sync.Mutex
	config      *TrackConfig
	updatedCh   chan struct{}
	updatedOnce sync.Once

	ctx context.Context
}

func (rss *receiveSubscribeStream) SubscribeID() SubscribeID {
	return rss.subscribeID
}

func (rss *receiveSubscribeStream) WriteInfo(info Info) error {
	var err error
	rss.acceptOnce.Do(func() {
		rss.configMu.Lock()
		defer rss.configMu.Unlock()
		if rss.ctx.Err() != nil {
			err = Cause(rss.ctx)
			return
		}
		sum := message.SubscribeOkMessage{
			GroupPeriod: message.GroupPeriod(info.GroupPeriod),
		}
		err = sum.Encode(rss.stream)
		if err != nil {
			rss.closeWithError(InternalSubscribeErrorCode)
			return
		}
	})

	return err
}

func (rss *receiveSubscribeStream) TrackConfig() *TrackConfig {
	rss.configMu.Lock()
	defer rss.configMu.Unlock()

	// Ensure config is never nil
	if rss.config == nil {
		rss.config = &TrackConfig{}
	}

	return rss.config
}

func (rss *receiveSubscribeStream) Updated() <-chan struct{} {
	return rss.updatedCh
}

func (rss *receiveSubscribeStream) close() error {
	rss.configMu.Lock()
	defer rss.configMu.Unlock()

	if rss.ctx.Err() != nil {
		return Cause(rss.ctx)
	}

	// Close the write-side stream
	err := rss.stream.Close()
	// Cancel the read-side stream
	rss.stream.CancelRead(quic.StreamErrorCode(PublishAbortedErrorCode))

	rss.updatedOnce.Do(func() {
		if rss.updatedCh != nil {
			close(rss.updatedCh)
		}
	})

	return err
}

func (rss *receiveSubscribeStream) closeWithError(code SubscribeErrorCode) error {
	if rss == nil {
		panic("receiveSubscribeStream: cannot call closeWithError on nil stream")
	}

	rss.configMu.Lock()
	defer rss.configMu.Unlock()

	if rss.ctx.Err() != nil {
		return Cause(rss.ctx)
	}

	strErrCode := quic.StreamErrorCode(code)
	// Cancel the write-side stream
	rss.stream.CancelWrite(strErrCode)
	// Cancel the read-side stream
	rss.stream.CancelRead(strErrCode)

	rss.updatedOnce.Do(func() {
		if rss.updatedCh != nil {
			close(rss.updatedCh)
		}
	})

	return nil
}
