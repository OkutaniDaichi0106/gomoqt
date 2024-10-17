package moqtransport

import (
	"reflect"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/internal/protocol"
)

func Test_getProtocolVersions(t *testing.T) {
	type args struct {
		versions []Version
	}
	tests := []struct {
		name string
		args args
		want []protocol.Version
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getProtocolVersions(tt.args.versions); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getProtocolVersions() = %v, want %v", got, tt.want)
			}
		})
	}
}
