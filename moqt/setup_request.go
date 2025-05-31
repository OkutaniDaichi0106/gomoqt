package moqt

import (
	"fmt"
	"net/url"
)

func NewSetupRequest(urlstr string) (*SetupRequest, error) {
	uri, err := url.Parse(urlstr)
	if err != nil {
		return nil, err
	}

	return &SetupRequest{
		URI:        uri,
		Parameters: nil,
	}, nil
}

type SetupRequest struct {
	Parameters *Parameters

	// URI is the URL of the server
	URI *url.URL
}

func (sr SetupRequest) String() string {
	if sr.Parameters == nil {
		return fmt.Sprintf("{ uri: %s, parameters: no parameters }", sr.URI)
	}
	return fmt.Sprintf("{ uri: %s, parameters: %s }", sr.URI, sr.Parameters.String())
}
