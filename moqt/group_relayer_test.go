package moqt_test

import (
	"testing"

	mock_transport "github.com/OkutaniDaichi0106/gomoqt/moqt/internal/transport/mock"
	mock_moqt "github.com/OkutaniDaichi0106/gomoqt/moqt/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewGroupBuffer(t *testing.T) {
	gr := &mock_moqt.MockGroupReader{
		frames: [][]byte{[]byte("frame1"), []byte("frame2")},
	}
	gb := NewGroupBuffer(gr)
	assert.NotNil(t, gb)
	assert.Equal(t, gr.GroupSequence(), gb.GroupSequence())
}

func TestGroupSequence(t *testing.T) {
	gr := &mockGroupReader{}
	gb := NewGroupBuffer(gr)
	assert.Equal(t, gr.GroupSequence(), gb.GroupSequence())
}

func TestClosed(t *testing.T) {
	gr := &mockGroupReader{}
	gb := NewGroupBuffer(gr)
	assert.False(t, gb.Closed())
	gr.closed = true
	assert.True(t, gb.Closed())
}

func TestRelay(t *testing.T) {
	gr := &mockGroupReader{
		frames: [][]byte{[]byte("frame1"), []byte("frame2")},
	}
	gb := NewGroupBuffer(gr)
	gw := &mockGroupWriter{}
	err := gb.Relay(gw)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(gw.writtenFrames))
	assert.Equal(t, []byte("frame1"), gw.writtenFrames[0])
	assert.Equal(t, []byte("frame2"), gw.writtenFrames[1])
}

func TestRelease(t *testing.T) {
	gr := &mockGroupReader{}
	gb := NewGroupBuffer(gr)
	gb.Release()
	assert.True(t, gb.Closed())
	assert.Empty(t, gb.bytes)
	assert.Empty(t, gb.frameRanges)
	assert.Nil(t, gb.err)
}

func TestRelayWithStream(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup test data
	gr := &mockGroupReader{
		frames: [][]byte{[]byte("frame1"), []byte("frame2")},
	}
	gb := NewGroupBuffer(gr)

	// Create mock stream and group writer
	ms := mock_transport.NewMockStream(ctrl)
	gw := &sendGroupStream{stream: ms}

	// Test successful relay
	err := gb.Relay(gw)
	assert.NoError(t, err)
	assert.Equal(t, append([]byte("frame1"), []byte("frame2")...), ms.written)

	// Test stream write error
	gr = &mockGroupReader{
		frames: [][]byte{[]byte("frame1"), []byte("frame2")},
	}
	gb = NewGroupBuffer(gr)
	ms = mock_transport.NewMockStream(ctrl)
	gw = &sendGroupStream{stream: ms}

	err = gb.Relay(gw)
	assert.Error(t, err)
	assert.Equal(t, "write error", err.Error())
}
