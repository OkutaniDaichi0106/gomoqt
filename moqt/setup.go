package moqt

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal"
	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
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
	supportedVersions []protocol.Version
	parsedURL         *url.URL
	once              bool
}

func (r SetupRequest) String() string {
	return fmt.Sprintf("SetupRequest: { URL: %s, SetupParameters: %s, supportedVersions: %v }",
		r.URL, r.SetupParameters.String(), r.supportedVersions)
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
	r.supportedVersions = internal.DefaultClientVersions

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
	Parameters Parameters

	selectedVersion protocol.Version
}

func (sr SetupResponce) String() string {
	return fmt.Sprintf("SetupResponce: { SelectedVersion: %d, Parameters: %s }", sr.selectedVersion, sr.Parameters.String())
}
