package moqt

import (
	"fmt"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/protocol"
)

type SetupResponse struct {
	Parameters *Parameters

	selectedVersion protocol.Version
}

func (sr SetupResponse) String() string {
	return fmt.Sprintf("{ selected_version: %d, parameters: %s }", sr.selectedVersion, sr.Parameters.String())
}
