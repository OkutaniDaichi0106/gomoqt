package moqtransport

import (
	"errors"
)

// func (a *Agent) listenControlChannel() chan error {
// 	errCh := make(chan error, 1)
// 	go func() {
// 		for data := range a.controlCh {
// 			switch MessageID(data[0]) {
// 			case SUBSCRIBE:
// 				// Check if the subscribe is acceptable
// 			case SUBSCRIBE_OK:
// 				// Send it to the Subscriber
// 				_, err := a.controlStream.Write(data)
// 				if err != nil {
// 					errCh <- err
// 					return
// 				}
// 			case SUBSCRIBE_ERROR:
// 				// Send it to the Subscriber
// 				_, err := a.controlStream.Write(data)
// 				if err != nil {
// 					errCh <- err
// 					return
// 				}
// 			case UNSUBSCRIBE:
// 				// Delete the Subscriber from the destinations
// 			default:
// 				errCh <- ErrUnexpectedMessage //TODO: handle the error as protocol violation
// 			}
// 		}
// 	}()
// 	return errCh
// }

var ErrProtocolViolation = errors.New("protocol violation")
