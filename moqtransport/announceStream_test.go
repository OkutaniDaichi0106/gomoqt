package moqtransport

import (
	"reflect"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/internal/moqtmessage"
)

func TestSendAnnounceStream_ReceiveSubscribeNamespace(t *testing.T) {
	tests := []struct {
		name    string
		sender  SendAnnounceStream
		want    moqtmessage.TrackNamespacePrefix
		want1   moqtmessage.Parameters
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := tt.sender.ReceiveSubscribeNamespace()
			if (err != nil) != tt.wantErr {
				t.Errorf("SendAnnounceStream.ReceiveSubscribeNamespace() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SendAnnounceStream.ReceiveSubscribeNamespace() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("SendAnnounceStream.ReceiveSubscribeNamespace() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestSendAnnounceStream_Announce(t *testing.T) {
	type args struct {
		trackNamespace moqtmessage.TrackNamespace
		config         AnnounceConfig
	}
	tests := []struct {
		name    string
		sender  SendAnnounceStream
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.sender.Announce(tt.args.trackNamespace, tt.args.config); (err != nil) != tt.wantErr {
				t.Errorf("SendAnnounceStream.Announce() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReceiveAnnounceStream_SubscribeNamespace(t *testing.T) {
	type args struct {
		trackNamespacePrefix moqtmessage.TrackNamespacePrefix
	}
	tests := []struct {
		name     string
		receiver ReceiveAnnounceStream
		args     args
		wantErr  bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.receiver.SubscribeNamespace(tt.args.trackNamespacePrefix); (err != nil) != tt.wantErr {
				t.Errorf("ReceiveAnnounceStream.SubscribeNamespace() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReceiveAnnounceStream_ReceiveAnnounce(t *testing.T) {
	tests := []struct {
		name     string
		receiver ReceiveAnnounceStream
		want     *Announcement
		wantErr  bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.receiver.ReceiveAnnounce()
			if (err != nil) != tt.wantErr {
				t.Errorf("ReceiveAnnounceStream.ReceiveAnnounce() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReceiveAnnounceStream.ReceiveAnnounce() = %v, want %v", got, tt.want)
			}
		})
	}
}
