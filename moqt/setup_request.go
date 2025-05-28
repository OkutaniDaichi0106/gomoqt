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
		uri:        uri,
		Parameters: nil,
	}, nil
}

type SetupRequest struct {
	Parameters *Parameters

	// URI is the URL of the server
	uri *url.URL
}

func (sr SetupRequest) String() string {
	if sr.Parameters == nil {
		return fmt.Sprintf("SetupRequest: { URI: %s, Parameters: No Parameters }", sr.uri)
	}
	return fmt.Sprintf("SetupRequest: { URI: %s, Parameters: %s }", sr.uri, sr.Parameters.String())
}
