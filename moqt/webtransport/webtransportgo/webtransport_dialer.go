package webtransportgo

import (
	"context"
	"errors"
	"net/http"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/quic"
	"github.com/quic-go/webtransport-go"
)

func Dial(ctx context.Context, addr string, header http.Header) (*http.Response, quic.Connection, error) {
	var d webtransport.Dialer
	rsp, wtsess, err := d.Dial(ctx, addr, header)
	if err != nil {
		return nil, nil, err
	}

	// Ensure wtsess is not nil before proceeding
	if wtsess == nil {
		err := errors.New("webtransport session is nil after dial")
		return nil, nil, err
	}

	return rsp, WrapWebTransportConnection(wtsess), nil
}
