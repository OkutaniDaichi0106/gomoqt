package moqtransport

import (
	"context"
	"io"
	"log"
	"sync"
	"time"

	"github.com/quic-go/webtransport-go"
)

type publisherAgenter interface {
	acceptAnnounce() error
	acceptObject() error
}

type PublisherAgent struct {
	clientAgent

	/*
	 * Destination the Agent send to
	 */
	destinations struct {
		sessions []*webtransport.Session
		mu       sync.Mutex
	}
}

func (*PublisherAgent) Role() Role {
	return PUB
}

/*
 * Handle announcement exchange
 * - Receive an ANNOUNCE message from the publisher
 * - Send ANNOUNCE_OK or ANNOUNCE_ERROR message to the publisher
 */
func AcceptAnnounce(agent publisherAgenter) error {
	return agent.acceptAnnounce()
}

func (a *PublisherAgent) acceptAnnounce() error {
	var err error
	var ae AnnounceError
	// Receive an ANNOUNCE message
	id, err := deserializeHeader(a.controlReader)
	if err != nil {
		return err
	}
	if id != ANNOUNCE {
		return ErrUnexpectedMessage
	}

	//TODO: handle error
	am := AnnounceMessage{}
	err = am.deserializeBody(a.controlReader)
	if err != nil {
		ae = AnnounceError{
			TrackNamespace: am.TrackNamespace,
			Code:           AnnounceErrorCode(ANNOUNCE_INTERNAL_ERROR),
			Reason:         ANNOUNCE_ERROR_REASON[ANNOUNCE_INTERNAL_ERROR],
		}
		_, err2 := a.controlStream.Write(ae.serialize()) // Handle the error when wrinting message
		log.Println(err2)

		return err
	} // TODO: handle the parameter

	// Register the ANNOUNCE message
	SERVER.announcements.add(am)

	//TODO
	_, ok := SERVER.publishers.index[am.TrackNamespace]
	if ok {
		ae = AnnounceError{
			TrackNamespace: am.TrackNamespace,
			Code:           DUPLICATE_TRACK_NAMESPACE,
			Reason:         ANNOUNCE_ERROR_REASON[DUPLICATE_TRACK_NAMESPACE],
		}
		_, err2 := a.controlStream.Write(ae.serialize()) // Handle the error when wrinting message
		log.Println(err2)

		return ErrDuplicatedNamespace
	}
	// Register the Publishers' Agent
	SERVER.publishers.add(am.TrackNamespace, a)

	// Send ANNOUNCE_OK message or ANNOUNCE_ERROR message as responce
	aom := AnnounceOkMessage{
		TrackNamespace: am.TrackNamespace,
	}
	_, err = a.controlStream.Write(aom.serialize())
	if err != nil {
		ae = AnnounceError{
			TrackNamespace: am.TrackNamespace,
			Code:           AnnounceErrorCode(ANNOUNCE_INTERNAL_ERROR),
			Reason:         ANNOUNCE_ERROR_REASON[ANNOUNCE_INTERNAL_ERROR],
		}
		_, err2 := a.controlStream.Write(ae.serialize()) // Handle the error when wrinting message
		log.Println(err2)

		return err
	}

	//
	log.Println("NOW:", SERVER.announcements.index)

	return nil
}

func AcceptObjects(agent *PublisherAgent, ctx context.Context) error {
	// Check the role
	if agent.role != PUB && agent.role != PUB_SUB {
		return ErrInvalidRole
	}
	return agent.acceptObjects(ctx)
}

func (a *PublisherAgent) acceptObjects(ctx context.Context) error {
	errCh := make(chan error, 1)

	for {
		// Accept a new unidirectional stream
		stream, err := a.session.AcceptUniStream(ctx)
		if err != nil {
			return err
		}
		log.Println("Accepted!!", stream)
		go func(stream webtransport.ReceiveStream) {
			buf := make([]byte, 1<<8)
			data := make([]byte, 0, 1<<8)
			for {
				n, err := stream.Read(buf)
				if err != nil {
					if err == io.EOF {
						// Append final data
						data = append(data, buf[:n]...)
						break
					}
					errCh <- err
					return
				}
				// Append read data
				data = append(data, buf[:n]...)
			}

			// Distribute the data to all Subscribers
			go a.distribute(data)

		}(stream)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case err = <-errCh:
			return err
		default:
			continue
		}
	}
}

func (a *PublisherAgent) distribute(b []byte) {
	var wg sync.WaitGroup
	for _, sess := range a.destinations.sessions {

		// Increment the wait group by 1 and count the number of current processes
		wg.Add(1)
		go func(sess *webtransport.Session) {
			defer wg.Done()
			// Set context to terminate the operation upon timeout
			ctx, cancel := context.WithTimeout(context.TODO(), 2*time.Second) // TODO: Left the duration to the user's implementation
			defer cancel()

			// Open a unidirectional stream
			stream, err := sess.OpenStreamSync(ctx)
			if err != nil {
				log.Println(err)
				return //TODO: handle the error
			}

			// Close the stream after whole data was sent
			defer stream.Close()

			// Send data on the stream
			_, err = stream.Write(b)
			if err != nil {
				log.Println(err)
				return //TODO: handle the error
			}
		}(sess)
	}

	// Wait untill the data has been sent to all sessions
	wg.Wait()
}
