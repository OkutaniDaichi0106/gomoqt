package moqtransport

import (
	"reflect"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/internal/moqtmessage"
)

func TestReceiveTrackStatusStream_RequestTrackStatus(t *testing.T) {
	type args struct {
		trackNamespace moqtmessage.TrackNamespace
		trackName      string
	}
	tests := []struct {
		name    string
		stream  *ReceiveTrackStatusStream
		args    args
		want    *TrackStatus
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.stream.RequestTrackStatus(tt.args.trackNamespace, tt.args.trackName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReceiveTrackStatusStream.RequestTrackStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReceiveTrackStatusStream.RequestTrackStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSendTrackStatusStream_ReceiveTrackStatusRequest(t *testing.T) {
	tests := []struct {
		name    string
		stream  SendTrackStatusStream
		want    *moqtmessage.TrackStatusRequestMessage
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.stream.ReceiveTrackStatusRequest()
			if (err != nil) != tt.wantErr {
				t.Errorf("SendTrackStatusStream.ReceiveTrackStatusRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SendTrackStatusStream.ReceiveTrackStatusRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSendTrackStatusStream_SendTrackStatus(t *testing.T) {
	type args struct {
		request moqtmessage.TrackStatusRequestMessage
		ts      TrackStatus
	}
	tests := []struct {
		name    string
		stream  SendTrackStatusStream
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.stream.SendTrackStatus(tt.args.request, tt.args.ts); (err != nil) != tt.wantErr {
				t.Errorf("SendTrackStatusStream.SendTrackStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
