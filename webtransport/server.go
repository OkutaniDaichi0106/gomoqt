package webtransport

import (
	"net/http"

	"github.com/OkutaniDaichi0106/gomoqt/webtransport/internal"
	"github.com/OkutaniDaichi0106/gomoqt/webtransport/webtransportgo"
)

func NewDefaultServer(checkOrigin func(*http.Request) bool) internal.Server {
	return webtransportgo.NewDefaultServer(checkOrigin)
}

type Server = internal.Server
