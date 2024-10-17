package moqtransport

import (
	"reflect"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/internal/moqtmessage"
	"github.com/quic-go/quic-go/quicvarint"
)

func TestSendSubscribeStream_Subscribe(t *testing.T) {
	type args struct {
		trackNamespace moqtmessage.TrackNamespace
		trackName      string
		config         SubscribeConfig
	}
	tests := []struct {
		name    string
		stream  SendSubscribeStream
		args    args
		want    *Subscription
		want1   *TrackStatus
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := tt.stream.Subscribe(tt.args.trackNamespace, tt.args.trackName, tt.args.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("SendSubscribeStream.Subscribe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SendSubscribeStream.Subscribe() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("SendSubscribeStream.Subscribe() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestSendSubscribeStream_UpdateSubscription(t *testing.T) {
	type args struct {
		subscription Subscription
		config       SubscribeConfig
	}
	tests := []struct {
		name    string
		stream  SendSubscribeStream
		args    args
		want    *TrackStatus
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.stream.UpdateSubscription(tt.args.subscription, tt.args.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("SendSubscribeStream.UpdateSubscription() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SendSubscribeStream.UpdateSubscription() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_receiveTrackStatus(t *testing.T) {
	type args struct {
		qvReader   quicvarint.Reader
		trackAlias moqtmessage.TrackAlias
	}
	tests := []struct {
		name    string
		args    args
		want    *TrackStatus
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := receiveTrackStatus(tt.args.qvReader, tt.args.trackAlias)
			if (err != nil) != tt.wantErr {
				t.Errorf("receiveTrackStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("receiveTrackStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSendSubscribeStream_CancelSubscribe(t *testing.T) {
	type args struct {
		err SubscribeError
	}
	tests := []struct {
		name   string
		stream SendSubscribeStream
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.stream.CancelSubscribe(tt.args.err)
		})
	}
}

func TestReceiveSubscribeStream_ReceiveSubscribe(t *testing.T) {
	tests := []struct {
		name    string
		stream  ReceiveSubscribeStream
		want    *Subscription
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.stream.ReceiveSubscribe()
			if (err != nil) != tt.wantErr {
				t.Errorf("ReceiveSubscribeStream.ReceiveSubscribe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReceiveSubscribeStream.ReceiveSubscribe() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReceiveSubscribeStream_AllowSubscribe(t *testing.T) {
	type args struct {
		subscription Subscription
		trackStatus  TrackStatus
	}
	tests := []struct {
		name    string
		stream  *ReceiveSubscribeStream
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.stream.AllowSubscribe(tt.args.subscription, tt.args.trackStatus); (err != nil) != tt.wantErr {
				t.Errorf("ReceiveSubscribeStream.AllowSubscribe() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReceiveSubscribeStream_RejectSubscribe(t *testing.T) {
	type args struct {
		err SubscribeError
	}
	tests := []struct {
		name   string
		stream ReceiveSubscribeStream
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.stream.RejectSubscribe(tt.args.err)
		})
	}
}

func TestReceiveSubscribeStream_ReceiveSubscribeUpdate(t *testing.T) {
	tests := []struct {
		name    string
		stream  ReceiveSubscribeStream
		want    *Subscription
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.stream.ReceiveSubscribeUpdate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ReceiveSubscribeStream.ReceiveSubscribeUpdate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReceiveSubscribeStream.ReceiveSubscribeUpdate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSubscription_GetConfig(t *testing.T) {
	tests := []struct {
		name string
		s    Subscription
		want SubscribeConfig
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.GetConfig(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Subscription.GetConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
