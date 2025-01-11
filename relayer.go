package moqt

import "sync"

// import "context"

// type Relayer interface {
// 	// Open Announce Stream to upstream
// 	OpenAnnounceStream(Interest) (ReceiveAnnounceStream, error)

// 	// Open Fetch Stream to upstream
// 	OpenFetchStream(Fetch) (ReceiveDataStream, error)

// 	// Open Info Stream to upstream
// 	OpenInfoStream(InfoRequest) (Info, error)

// 	// Open Subscribe Stream to upstream
// 	OpenSubscribeStream(Subscription) (SendSubscribeStream, error)

// 	//
// 	ListenAndServe(ctx context.Context) error
// }

// func NewRelayer(manager *relayManager, sess ServerSession) Relayer {
// 	return nil
// }

// var _ Relayer = (*relayer)(nil)

type relayer struct {
	manager   RelayManager
	taskQueue chan func()
}

type task struct {
}

type taskQueue struct {
	tasks []*task
	mu    sync.Mutex
	cond  *sync.Cond
}

// func (r *relayer) ListenAndServe(ctx context.Context) error {

// 	go r.listenSubscribeStream(ctx)

// 	go r.listenAnnounceStream(ctx)

// 	go r.listenFetchStream(ctx)

// 	go r.listenInfoStream(ctx)

// 	return nil
// }
// func (r *relayer) listenAnnounceStream(ctx context.Context) {
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return
// 		default:
// 			// Accept a new stream
// 			stream, err := r.sess.AcceptAnnounceStream(ctx)
// 			if err != nil {
// 				return
// 			}

// 			r.manager.HandleAnnounceStream(stream)
// 		}
// 	}

// }
// func (r *relayer) listenSubscribeStream(ctx context.Context) {
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return
// 		default:
// 			// Accept a new stream
// 			stream, err := r.sess.AcceptSubscribeStream(ctx)
// 			if err != nil {
// 				return
// 			}

// 			r.manager.HandleSubscribeStream(stream)
// 		}
// 	}

// }

// func (r *relayer) listenFetchStream(ctx context.Context) {
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return
// 		default:
// 			// Accept a new stream
// 			stream, err := r.sess.AcceptFetchStream(ctx)
// 			if err != nil {
// 				return
// 			}

// 			r.manager.HandleFetchStream(stream)

// 		}
// 	}

// }

// func (r *relayer) listenInfoStream(ctx context.Context) {
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return
// 		default:
// 			// Accept a new stream
// 			stream, err := r.sess.AcceptInfoStream(ctx)
// 			if err != nil {
// 				return
// 			}

// 			r.manager.HandleInfoStream(stream)
// 		}
// 	}

// }
