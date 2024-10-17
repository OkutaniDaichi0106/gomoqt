package moqtransport

import (
	"reflect"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/internal/moqtmessage"
)

func Test_trackAliasMap_getAlias(t *testing.T) {
	type args struct {
		tns moqtmessage.TrackNamespace
		tn  string
	}
	tests := []struct {
		name  string
		tamap *trackAliasMap
		args  args
		want  moqtmessage.TrackAlias
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tamap.getAlias(tt.args.tns, tt.args.tn); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("trackAliasMap.getAlias() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_trackAliasMap_getName(t *testing.T) {
	type args struct {
		ta moqtmessage.TrackAlias
	}
	tests := []struct {
		name  string
		tamap *trackAliasMap
		args  args
		want  moqtmessage.TrackNamespace
		want1 string
		want2 bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := tt.tamap.getName(tt.args.ta)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("trackAliasMap.getName() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("trackAliasMap.getName() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("trackAliasMap.getName() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}
