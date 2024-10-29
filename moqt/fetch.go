package moqt

import "github.com/quic-go/quic-go/quicvarint"

type FetchStream Stream

type FetchHandler interface {
	HandleFetch(FetchRequest, FetchResponceWriter)
}

type FetchRequest struct{}

type FetchRequestReader interface {
	Read(quicvarint.Reader) (FetchRequest, error)
}

type FetchResponceWriter interface {
}
