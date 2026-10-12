package webtransportgo

import (
	"context"
	"testing"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDial_InvalidAddress(t *testing.T) {
	rsp, conn, err := Dial(context.Background(), "://bad-url", nil, nil, nil, nil)

	require.Error(t, err)
	assert.Nil(t, rsp)
	assert.Nil(t, conn)
}

func TestDialQUICConfig_Nil(t *testing.T) {
	assert.Nil(t, dialQUICConfig(nil), "nil leaves the library its defaults")
}

func TestDialQUICConfig(t *testing.T) {
	given := &quic.Config{MaxIdleTimeout: time.Minute, InitialPacketSize: 1200}

	got := dialQUICConfig(given)

	assert.Equal(t, &quic.Config{
		MaxIdleTimeout:                   time.Minute,
		InitialPacketSize:                1200,
		EnableDatagrams:                  true,
		EnableStreamResetPartialDelivery: true,
	}, got, "the caller's settings, with what WebTransport needs turned on")
	assert.Equal(t, &quic.Config{MaxIdleTimeout: time.Minute, InitialPacketSize: 1200}, given, "the caller's configuration is left alone")
}
