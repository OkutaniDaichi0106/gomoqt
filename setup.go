package moqt

import (
	"errors"
	"io"
	"log/slog"
	"net/url"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
)

// func NewSetupRequest(urlstr string, params Parameters) (req SetupRequest, err error) {
// 	// Parse the URI
// 	parsedURL, err := url.ParseRequestURI(urlstr)
// 	if err != nil {
// 		slog.Error("failed to parse the url", slog.String("error", err.Error()))
// 		return req, err
// 	}

// 	if parsedURL.Scheme != "https" && parsedURL.Scheme != "moqt" {
// 		return req, errors.New("invalid url scheme. moqt or https scheme is required")
// 	}

// 	return SetupRequest{
// 		supportedVersions: DefaultClientVersions,
// 		URL:               urlstr,
// 		parsedURL:         parsedURL,
// 		MaxSubscribeID:    0, // TODO:
// 		SetupParameters:   params,
// 	}, nil
// }

/*
 *
 */
type SetupRequest struct {
	// Required
	URL string

	// Options
	MaxSubscribeID  uint64 // TODO:
	SetupParameters Parameters

	// Internal
	supportedVersions []Version
	parsedURL         *url.URL
	once              bool
}

func (r *SetupRequest) init() error {
	if r.once {
		return nil
	}

	if r.URL == "" {
		return errors.New("URL is required")
	}

	// Parse the URI
	parsedURL, err := url.ParseRequestURI(r.URL)
	if err != nil {
		slog.Error("failed to parse the url", slog.String("error", err.Error()))
		return err
	}
	// Check the scheme
	if parsedURL.Scheme != "https" && parsedURL.Scheme != "moqt" {
		return errors.New("invalid url scheme. moqt or https scheme is required")
	}

	// Set the parsed URL
	r.parsedURL = parsedURL

	// Set the Versions
	r.supportedVersions = DefaultClientVersions

	// Initialize the SetupParameters
	if r.SetupParameters == nil {
		r.SetupParameters = make(Parameters)
	}
	if parsedURL.Scheme == "moqt" {
		// Set up the path parameter if the scheme is "moqt"
		r.SetupParameters.Add(PATH, r.parsedURL.Path)
	}

	// Set the flag
	r.once = true

	return nil
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
