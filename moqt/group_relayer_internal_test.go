package moqt

import (
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/transport"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func newMockSendGroupStream(id SubscribeID, seq GroupSequence, stream transport.MockSendStream) SendGroupStream {
	return &sendGroupStream{
		subscribeID: id,
		sequence:    seq,
		stream:      &stream,
	}
}

func TestGroupRelayer_SendGroupStream(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	stream := transport.NewMockSendStream(ctrl)
	stream.EXPECT().Write(gomock.Any()).Return(0, nil).Times(1)

	sgs := newMockSendGroupStream(0, 0, *stream)
	err := sgs.WriteFrame([]byte{0x00, 0x01, 0x02, 0x03})
	assert.Nil(t, err, "WriteFrame should not return an error")
}
