package moqt

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewTrackSender(t *testing.T) {
	mockStream := &MockQUICStream{}
	substr := newReceiveSubscribeStream(SubscribeID(1), mockStream, &SubscribeConfig{})

	openGroupFunc := func(seq GroupSequence) (*sendGroupStream, error) {
		return &sendGroupStream{}, nil
	}

	sender := newTrackSender(substr, openGroupFunc)

	assert.NotNil(t, sender)
	assert.NotNil(t, sender.openGroupFunc)
	assert.Equal(t, substr, sender.subscribeStream)
}

func TestTrackSender_OpenGroup(t *testing.T) {
	tests := map[string]struct {
		subscribeStream *receiveSubscribeStream
		openGroupFunc   func(GroupSequence) (*sendGroupStream, error)
		seq             GroupSequence
		wantErr         bool
	}{
		"success": {
			subscribeStream: func() *receiveSubscribeStream {
				mockStream := &MockQUICStream{
					ReadFunc: func(p []byte) (int, error) {
						// Block indefinitely to keep stream open
						select {}
					},
					WriteFunc: func(p []byte) (int, error) {
						return len(p), nil
					},
				}
				mockStream.On("Read", mock.AnythingOfType("[]uint8"))
				mockStream.On("Write", mock.AnythingOfType("[]uint8"))
				mockStream.On("Close").Return(nil)
				mockStream.On("CancelRead", mock.AnythingOfType("quic.StreamErrorCode")).Return()
				mockStream.On("StreamID").Return(quic.StreamID(1))
				return newReceiveSubscribeStream(SubscribeID(1), mockStream, &SubscribeConfig{})
			}(),
			openGroupFunc: func(seq GroupSequence) (*sendGroupStream, error) {
				return newSendGroupStream(&MockQUICSendStream{}, seq), nil
			},
			seq:     GroupSequence(1),
			wantErr: false,
		},
		"subscribe stream closed": {
			subscribeStream: func() *receiveSubscribeStream {
				mockStream := &MockQUICStream{
					ReadFunc: func(p []byte) (int, error) {
						return 0, io.EOF
					},
					WriteFunc: func(p []byte) (int, error) {
						return len(p), nil
					},
				}
				mockStream.On("Read", mock.AnythingOfType("[]uint8"))
				mockStream.On("Write", mock.AnythingOfType("[]uint8"))
				mockStream.On("Close").Return(nil)
				mockStream.On("CancelRead", mock.AnythingOfType("quic.StreamErrorCode")).Return()
				mockStream.On("StreamID").Return(quic.StreamID(1))
				substr := newReceiveSubscribeStream(SubscribeID(1), mockStream, &SubscribeConfig{})
				substr.close()
				return substr
			}(),
			openGroupFunc: func(seq GroupSequence) (*sendGroupStream, error) {
				return &sendGroupStream{closedCh: make(chan struct{})}, nil
			},
			seq:     GroupSequence(2),
			wantErr: true,
		},
		"open group error": {
			subscribeStream: func() *receiveSubscribeStream {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				mockStream := &MockQUICStream{
					ReadFunc: func(p []byte) (int, error) {
						<-ctx.Done()
						return 0, io.EOF
					},
					WriteFunc: func(p []byte) (int, error) {
						return len(p), nil
					},
				}
				mockStream.On("Read", mock.AnythingOfType("[]uint8"))
				mockStream.On("Write", mock.AnythingOfType("[]uint8"))
				mockStream.On("Close").Return(nil)
				mockStream.On("CancelRead", mock.AnythingOfType("quic.StreamErrorCode")).Return()
				mockStream.On("StreamID").Return(quic.StreamID(1))
				return newReceiveSubscribeStream(SubscribeID(1), mockStream, &SubscribeConfig{})
			}(),
			openGroupFunc: func(seq GroupSequence) (*sendGroupStream, error) {
				return nil, errors.New("mock error")
			},
			seq:     GroupSequence(3),
			wantErr: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			substr := tt.subscribeStream
			openGroupFunc := tt.openGroupFunc
			sender := newTrackSender(substr, openGroupFunc)

			groupWriter, err := sender.OpenGroup(tt.seq)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, groupWriter)

				sender.mu.Lock()
				queueSize := len(sender.queue)
				sender.mu.Unlock()
				assert.Equal(t, 0, queueSize)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, groupWriter)

				sender.mu.Lock()
				queueSize := len(sender.queue)
				sender.mu.Unlock()
				assert.Equal(t, 1, queueSize)
			}
		})
	}
}

func TestTrackSender_Close(t *testing.T) {
	mockStream := &MockQUICStream{
		ReadFunc: func(p []byte) (int, error) {
			select {}
		},
		WriteFunc: func(p []byte) (int, error) {
			return len(p), nil
		},
	}
	mockStream.On("Read", mock.AnythingOfType("[]uint8"))
	mockStream.On("Write", mock.AnythingOfType("[]uint8"))
	mockStream.On("Close").Return(nil)
	mockStream.On("CancelRead", mock.AnythingOfType("quic.StreamErrorCode")).Return()
	mockStream.On("CancelWrite", mock.AnythingOfType("quic.StreamErrorCode")).Return()
	mockStream.On("StreamID").Return(quic.StreamID(1))

	substr := newReceiveSubscribeStream(SubscribeID(1), mockStream, &SubscribeConfig{})

	mockSendStream := &MockQUICSendStream{}
	mockSendStream.On("Close").Return(nil)
	mockSendStream.On("StreamID").Return(quic.StreamID(1))

	testStream := newSendGroupStream(mockSendStream, GroupSequence(1))

	openGroupFunc := func(seq GroupSequence) (*sendGroupStream, error) {
		return testStream, nil
	}

	sender := newTrackSender(substr, openGroupFunc)

	_, err := sender.OpenGroup(GroupSequence(1))
	assert.NoError(t, err)

	sender.mu.Lock()
	initialQueueSize := len(sender.queue)
	sender.mu.Unlock()
	assert.Equal(t, 1, initialQueueSize)

	err = sender.Close()

	assert.NoError(t, err)

	sender.mu.Lock()
	finalQueueSize := len(sender.queue)
	sender.mu.Unlock()
	assert.Equal(t, 0, finalQueueSize)
}

func TestTrackSender_Close_AlreadyClosed(t *testing.T) {
	mockStream := &MockQUICStream{}
	mockStream.On("Read", mock.Anything).Return(0, io.EOF)
	mockStream.On("Write", mock.AnythingOfType("[]uint8")).Return(func(p []byte) int { return len(p) }, nil)
	mockStream.On("Close").Return(nil)
	mockStream.On("CancelRead", mock.AnythingOfType("quic.StreamErrorCode")).Return()
	mockStream.On("CancelWrite", mock.AnythingOfType("quic.StreamErrorCode")).Return()
	mockStream.On("StreamID").Return(quic.StreamID(1))

	substr := newReceiveSubscribeStream(SubscribeID(1), mockStream, &SubscribeConfig{})
	substr.close()

	openGroupFunc := func(seq GroupSequence) (*sendGroupStream, error) {
		mockSendStream := &MockQUICSendStream{}
		mockSendStream.On("Close").Return(nil)
		mockSendStream.On("StreamID").Return(quic.StreamID(1))
		return newSendGroupStream(mockSendStream, GroupSequence(1)), nil
	}

	sender := newTrackSender(substr, openGroupFunc)

	err := sender.Close()

	assert.NoError(t, err)

	sender.mu.Lock()
	finalQueueSize := len(sender.queue)
	sender.mu.Unlock()
	assert.Equal(t, 0, finalQueueSize)
}

func TestTrackSender_CloseWithError(t *testing.T) {
	tests := map[string]struct {
		errorCode    SubscribeErrorCode
		streamClosed bool
	}{
		"custom error": {
			errorCode:    SubscribeErrorCode(1),
			streamClosed: false,
		},
		"zero error": {
			errorCode:    SubscribeErrorCode(0),
			streamClosed: false,
		},
		"already closed": {
			errorCode:    SubscribeErrorCode(2),
			streamClosed: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockStream := &MockQUICStream{
				ReadFunc: func(p []byte) (int, error) {
					select {}
				},
				WriteFunc: func(p []byte) (int, error) {
					return len(p), nil
				},
			}
			mockStream.On("Read", mock.AnythingOfType("[]uint8"))
			mockStream.On("Write", mock.AnythingOfType("[]uint8"))
			mockStream.On("Close").Return(nil)
			mockStream.On("CancelRead", mock.AnythingOfType("quic.StreamErrorCode")).Return()
			mockStream.On("CancelWrite", mock.AnythingOfType("quic.StreamErrorCode")).Return()
			mockStream.On("StreamID").Return(quic.StreamID(1))

			substr := newReceiveSubscribeStream(SubscribeID(1), mockStream, &SubscribeConfig{})

			if tt.streamClosed {
				substr.close()
			}

			mockSendStream := &MockQUICSendStream{}
			mockSendStream.On("StreamID").Return(quic.StreamID(1))
			mockSendStream.On("CancelWrite", mock.AnythingOfType("quic.StreamErrorCode")).Return()

			openGroupFunc := func(seq GroupSequence) (*sendGroupStream, error) {
				return newSendGroupStream(mockSendStream, seq), nil
			}

			sender := newTrackSender(substr, openGroupFunc)

			if !tt.streamClosed {
				_, err := sender.OpenGroup(GroupSequence(1))
				assert.NoError(t, err)

				sender.mu.Lock()
				initialQueueSize := len(sender.queue)
				sender.mu.Unlock()
				assert.Equal(t, 1, initialQueueSize)
			}

			err := sender.CloseWithError(tt.errorCode)

			assert.NoError(t, err)

			sender.mu.Lock()
			finalQueueSize := len(sender.queue)
			sender.mu.Unlock()
			assert.Equal(t, 0, finalQueueSize)
		})
	}
}

func TestTrackSender_ConcurrentOperations(t *testing.T) {
	mockStream := &MockQUICStream{
		ReadFunc: func(p []byte) (int, error) {
			select {}
		},
		WriteFunc: func(p []byte) (int, error) {
			return len(p), nil
		},
	}

	mockStream.On("Read", mock.AnythingOfType("[]uint8"))
	mockStream.On("Write", mock.AnythingOfType("[]uint8"))
	mockStream.On("Close").Return(nil)
	mockStream.On("CancelRead", mock.AnythingOfType("quic.StreamErrorCode")).Return()
	mockStream.On("CancelWrite", mock.AnythingOfType("quic.StreamErrorCode")).Return()
	mockStream.On("StreamID").Return(quic.StreamID(1))

	substr := newReceiveSubscribeStream(SubscribeID(1), mockStream, &SubscribeConfig{})

	openGroupFunc := func(seq GroupSequence) (*sendGroupStream, error) {
		mockSendStream := &MockQUICSendStream{}
		mockSendStream.On("Close").Return(nil)
		mockSendStream.On("StreamID").Return(quic.StreamID(int(seq)))
		return newSendGroupStream(mockSendStream, seq), nil
	}

	sender := newTrackSender(substr, openGroupFunc)

	const numGroups = 10
	errChan := make(chan error, numGroups)

	for i := range numGroups {
		go func(seq GroupSequence) {
			_, err := sender.OpenGroup(seq)
			errChan <- err
		}(GroupSequence(i + 1))
	}

	for range numGroups {
		err := <-errChan
		assert.NoError(t, err)
	}

	sender.mu.Lock()
	queueSize := len(sender.queue)
	sender.mu.Unlock()
	assert.Equal(t, numGroups, queueSize)

	err := sender.Close()
	assert.NoError(t, err)

	sender.mu.Lock()
	finalQueueSize := len(sender.queue)
	sender.mu.Unlock()
	assert.Equal(t, 0, finalQueueSize)
}
