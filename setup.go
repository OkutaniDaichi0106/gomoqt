package moqt

import (
	"errors"
	"io"
	"log/slog"
	"net/url"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
)

func NewSetupRequest(urlstr string, params Parameters) (req SetupRequest, err error) {
	// Parse the URI
	parsedURL, err := url.ParseRequestURI(urlstr)
	if err != nil {
		slog.Error("failed to parse the url", slog.String("error", err.Error()))
		return req, err
	}

	if parsedURL.Scheme != "https" && parsedURL.Scheme != "moqt" {
		return req, errors.New("invalid url scheme. moqt or https scheme is required")
	}

	return SetupRequest{
		supportedVersions: DefaultClientVersions,
		urlstr:            urlstr,
		parsedURL:         parsedURL,
		MaxSubscribeID:    0, // TODO:
		Parameters:        params,
	}, nil
}

/*
 *
 */
type SetupRequest struct {
	supportedVersions []Version
	urlstr            string
	parsedURL         *url.URL
	MaxSubscribeID    uint64 // TODO:
	Parameters        Parameters
}

/*
 * Server
 */
type SetupResponce struct {
	SelectedVersion Version
	Parameters      Parameters
}

func readSetupResponce(r io.Reader) (SetupResponce, error) {
	/***/
	var ssm message.SessionServerMessage
	err := ssm.Decode(r)
	if err != nil {
		slog.Error("failed to read a SESSION_SERVER message", slog.String("error", err.Error()))
		return SetupResponce{}, err
	}

	return SetupResponce{
		SelectedVersion: Version(ssm.SelectedVersion),
		Parameters:      Parameters(ssm.Parameters),
	}, nil
}
