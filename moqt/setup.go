package moqt

import (
	"fmt"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
)

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
