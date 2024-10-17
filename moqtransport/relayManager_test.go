package moqtransport

import (
	"reflect"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/internal/moqtmessage"
)

func TestTrackManager_newTrackNamespace(t *testing.T) {
	type args struct {
		trackNamespace moqtmessage.TrackNamespace
	}
	tests := []struct {
		name string
		tm   *TrackManager
		args args
		want *trackNamespaceNode
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tm.newTrackNamespace(tt.args.trackNamespace); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TrackManager.newTrackNamespace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrackManager_findTrackNamespace(t *testing.T) {
	type args struct {
		trackNamespace moqtmessage.TrackNamespace
	}
	tests := []struct {
		name  string
		tm    *TrackManager
		args  args
		want  *trackNamespaceNode
		want1 bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := tt.tm.findTrackNamespace(tt.args.trackNamespace)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TrackManager.findTrackNamespace() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("TrackManager.findTrackNamespace() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestTrackManager_removeTrackNamespace(t *testing.T) {
	type args struct {
		trackNamespace moqtmessage.TrackNamespace
	}
	tests := []struct {
		name    string
		tm      *TrackManager
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.tm.removeTrackNamespace(tt.args.trackNamespace); (err != nil) != tt.wantErr {
				t.Errorf("TrackManager.removeTrackNamespace() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTrackManager_findTrack(t *testing.T) {
	type args struct {
		trackNamespace moqtmessage.TrackNamespace
		trackName      string
	}
	tests := []struct {
		name  string
		tm    *TrackManager
		args  args
		want  *trackNameNode
		want1 bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := tt.tm.findTrack(tt.args.trackNamespace, tt.args.trackName)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TrackManager.findTrack() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("TrackManager.findTrack() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_newTrackNamespaceTree(t *testing.T) {
	tests := []struct {
		name string
		want *TrackNamespaceTree
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newTrackNamespaceTree(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newTrackNamespaceTree() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrackNamespaceTree_insert(t *testing.T) {
	type args struct {
		tns moqtmessage.TrackNamespace
	}
	tests := []struct {
		name string
		tree TrackNamespaceTree
		args args
		want *trackNamespaceNode
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tree.insert(tt.args.tns); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TrackNamespaceTree.insert() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrackNamespaceTree_remove(t *testing.T) {
	type args struct {
		tns moqtmessage.TrackNamespace
	}
	tests := []struct {
		name    string
		tree    TrackNamespaceTree
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.tree.remove(tt.args.tns); (err != nil) != tt.wantErr {
				t.Errorf("TrackNamespaceTree.remove() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTrackNamespaceTree_trace(t *testing.T) {
	type args struct {
		tns moqtmessage.TrackNamespace
	}
	tests := []struct {
		name  string
		tree  TrackNamespaceTree
		args  args
		want  *trackNamespaceNode
		want1 bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := tt.tree.trace(tt.args.tns)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TrackNamespaceTree.trace() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("TrackNamespaceTree.trace() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_trackNamespaceNode_remove(t *testing.T) {
	type args struct {
		tns   moqtmessage.TrackNamespace
		depth int
	}
	tests := []struct {
		name    string
		node    *trackNamespaceNode
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.node.remove(tt.args.tns, tt.args.depth)
			if (err != nil) != tt.wantErr {
				t.Errorf("trackNamespaceNode.remove() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("trackNamespaceNode.remove() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_trackNamespaceNode_trace(t *testing.T) {
	type args struct {
		values []string
	}
	tests := []struct {
		name  string
		node  *trackNamespaceNode
		args  args
		want  *trackNamespaceNode
		want1 bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := tt.node.trace(tt.args.values...)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("trackNamespaceNode.trace() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("trackNamespaceNode.trace() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_trackNamespaceNode_findTrackName(t *testing.T) {
	type args struct {
		trackName string
	}
	tests := []struct {
		name  string
		node  *trackNamespaceNode
		args  args
		want  *trackNameNode
		want1 bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := tt.node.findTrackName(tt.args.trackName)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("trackNamespaceNode.findTrackName() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("trackNamespaceNode.findTrackName() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_trackNamespaceNode_newTrackNameNode(t *testing.T) {
	type args struct {
		trackName string
	}
	tests := []struct {
		name string
		node *trackNamespaceNode
		args args
		want *trackNameNode
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.node.newTrackNameNode(tt.args.trackName); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("trackNamespaceNode.newTrackNameNode() = %v, want %v", got, tt.want)
			}
		})
	}
}
