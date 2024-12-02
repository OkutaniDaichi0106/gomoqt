package moqt

import (
	"io"
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/OkutaniDaichi0106/gomoqt/internal/moq"
)

type SessionStream moq.Stream //TODO:

/*
 *
 */
type SetupRequest struct {
	SupportedVersions []Version
	Path              string // TODO:
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
