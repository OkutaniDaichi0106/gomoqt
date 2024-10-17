package moqtransport

import "github.com/OkutaniDaichi0106/gomoqt/moqtransport/internal/protocol"

type Version uint64

const (
	FoalkDraft01 Version = Version(protocol.FoalkDraft01)
)

func getProtocolVersions(versions []Version) []protocol.Version {
	pvs := make([]protocol.Version, 0)
	for _, v := range versions {
		pvs = append(pvs, protocol.Version(v))
	}
	return pvs
}
