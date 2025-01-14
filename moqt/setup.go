package moqt

import (
	"errors"
	"io"
	"log/slog"
	"net/url"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
)

/*
 *
 */
type SetupRequest struct {
	// Required
	URL string

	// Options
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
	if r.SetupParameters.paramMap == nil {
		r.SetupParameters = NewParameters()
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
	slog.Debug("reading a set-up responce")
	/***/
	var ssm message.SessionServerMessage
	err := ssm.Decode(r)
	if err != nil {
		slog.Error("failed to read a SESSION_SERVER message", slog.String("error", err.Error()))
		return SetupResponce{}, err
	}

	slog.Debug("read a set-up responce")

	return SetupResponce{
		SelectedVersion: Version(ssm.SelectedVersion),
		Parameters:      Parameters{ssm.Parameters},
	}, nil
}
