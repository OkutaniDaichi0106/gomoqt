package moqt

import (
	"bytes"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
)

type AnnouncementWriter interface {
	SendAnnouncement(announcement *Announcement) error
	Close() error
	CloseWithError(code AnnounceErrorCode) error
}

// func (w AnnouncementWriter) SendAnnouncements(announcements []*Announcement) error {}

func newSendAnnounceStream(stream quic.Stream, prefix string) *sendAnnounceStream {
	sas := &sendAnnounceStream{
		prefix:          prefix,
		stream:          stream,
		actives:         make(map[string]*Announcement),
		pendings:        make(map[string]message.AnnounceMessage),
		sendCh:          make(chan struct{}, 1),
		batchTimer:      time.NewTimer(100 * time.Millisecond), // Default batch timeout
		batchTimeout:    100 * time.Millisecond,
		processingTasks: make(map[string]*sync.WaitGroup),
	}

	// Stop the timer initially
	if !sas.batchTimer.Stop() {
		<-sas.batchTimer.C
	}

	go func() {
		for range sas.sendCh {
			err := sas.send()
			if err != nil {
				slog.Error("failed to send announcements", "err", err)
			}
		}
	}()

	return sas
}

var _ AnnouncementWriter = (*sendAnnounceStream)(nil)

type sendAnnounceStream struct {
	prefix string

	stream quic.Stream

	mu sync.Mutex

	actives    map[string]*Announcement
	pendings   map[string]message.AnnounceMessage
	pendingsMu sync.Mutex

	// Batch processing fields
	batchTimer      *time.Timer
	batchTimeout    time.Duration
	processingTasks map[string]*sync.WaitGroup
	tasksMu         sync.Mutex

	closed   bool
	closeErr error

	sendCh chan struct{} // Channel to trigger sending announcements
}

func (sas *sendAnnounceStream) SendAnnouncement(announcement *Announcement) error {
	sas.mu.Lock()
	if sas.closed {
		sas.mu.Unlock()
		if sas.closeErr != nil {
			return sas.closeErr
		}
		return errors.New("stream already closed")
	}
	sas.mu.Unlock()

	// Get suffix for this announcement
	suffix, ok := announcement.BroadcastPath().GetSuffix(sas.prefix)
	if !ok {
		return errors.New("invalid broadcast path")
	}

	// Create a WaitGroup for this batch of tasks
	var taskWg sync.WaitGroup

	sas.tasksMu.Lock()
	sas.processingTasks[suffix] = &taskWg
	sas.tasksMu.Unlock()

	taskWg.Add(1)

	go func(announcement *Announcement) {
		defer taskWg.Done()

		sas.pendingsMu.Lock()
		if active, ok := sas.actives[suffix]; ok {
			active.cancel()
			delete(sas.actives, suffix)
		}

		sas.actives[suffix] = announcement
		sas.pendingsMu.Unlock()

		sas.set(suffix, true)

		<-announcement.AwaitEnd()

		taskWg.Add(1)
		go func() {
			defer taskWg.Done()
			sas.set(suffix, false)

			sas.pendingsMu.Lock()
			delete(sas.actives, suffix)
			sas.pendingsMu.Unlock()
		}()
	}(announcement)

	// Start a goroutine to wait for all tasks for this suffix to complete
	go func() {
		taskWg.Wait()

		sas.tasksMu.Lock()
		delete(sas.processingTasks, suffix)
		allTasksComplete := (len(sas.processingTasks) == 0)
		sas.tasksMu.Unlock()

		if allTasksComplete {
			select {
			case sas.sendCh <- struct{}{}:
				// Successfully triggered send
			default:
				// Channel is full, send is already pending
			}
		}
	}()

	return nil
}

func (sas *sendAnnounceStream) set(suffix string, active bool) {
	sas.pendingsMu.Lock()
	defer sas.pendingsMu.Unlock()

	_, ok := sas.pendings[suffix]

	if active {
		sas.pendings[suffix] = message.AnnounceMessage{
			AnnounceStatus: message.ACTIVE,
			TrackSuffix:    suffix,
		}
	} else {
		if ok {
			delete(sas.pendings, suffix)
		} else {
			sas.pendings[suffix] = message.AnnounceMessage{
				AnnounceStatus: message.ENDED,
				TrackSuffix:    suffix,
			}
		}
	}
}

func (sas *sendAnnounceStream) send() error {
	sas.mu.Lock()
	defer sas.mu.Unlock()

	if sas.closed {
		if sas.closeErr != nil {
			return sas.closeErr
		}
		return errors.New("stream already closed")
	}

	sas.pendingsMu.Lock()
	if len(sas.pendings) == 0 {
		sas.pendingsMu.Unlock()
		return nil
	}

	// Calculate total length for buffer allocation
	var totalLen int
	for _, am := range sas.pendings {
		totalLen += am.Len()
	}

	buf := bytes.NewBuffer(make([]byte, 0, totalLen))

	// Encode all pending messages
	for _, am := range sas.pendings {
		am.Encode(buf)
	}

	// Clear pendings after encoding
	sas.pendings = make(map[string]message.AnnounceMessage)
	sas.pendingsMu.Unlock()

	// Write to stream
	_, err := sas.stream.Write(buf.Bytes())
	if err != nil {
		var strErr *quic.StreamError
		if errors.As(err, &strErr) {
			sas.closed = true
			sas.closeErr = &AnnounceError{
				StreamError: strErr,
			}

			return &AnnounceError{
				StreamError: strErr,
			}
		}

		return err
	}

	return nil
}

func (sas *sendAnnounceStream) Close() error {
	sas.mu.Lock()
	defer sas.mu.Unlock()
	if sas.closed {
		if sas.closeErr != nil {
			return sas.closeErr
		}
		return nil
	}

	sas.closed = true

	close(sas.sendCh)

	err := sas.stream.Close()
	if err != nil {
		return err
	}

	return nil
}

func (sas *sendAnnounceStream) CloseWithError(code AnnounceErrorCode) error {
	sas.mu.Lock()
	defer sas.mu.Unlock()

	if sas.closed {
		if sas.closeErr != nil {
			return sas.closeErr
		}
		return nil
	}

	sas.closed = true

	strErrCode := quic.StreamErrorCode(code)
	sas.closeErr = &AnnounceError{
		StreamError: &quic.StreamError{
			StreamID:  sas.stream.StreamID(),
			ErrorCode: strErrCode,
		},
	}

	close(sas.sendCh)

	sas.stream.CancelWrite(strErrCode)
	sas.stream.CancelRead(strErrCode)

	return nil
}
