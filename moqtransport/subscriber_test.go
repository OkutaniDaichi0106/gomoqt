package moqtransport

import (
	"context"
	"io"
	"reflect"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/internal/moqtmessage"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/webtransport-go"
)

func TestSubscriber_SetupMORQ(t *testing.T) {
	type args struct {
		qconn quic.Connection
		path  string
	}
	tests := []struct {
		name    string
		s       *Subscriber
		args    args
		want    *Session
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.SetupMORQ(tt.args.qconn, tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Subscriber.SetupMORQ() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Subscriber.SetupMORQ() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSubscriber_setupMORQ(t *testing.T) {
	type args struct {
		conn Connection
		path string
	}
	tests := []struct {
		name    string
		s       *Subscriber
		args    args
		want    *Session
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.setupMORQ(tt.args.conn, tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Subscriber.setupMORQ() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Subscriber.setupMORQ() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSubscriber_ReceiveDatagram(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		s       Subscriber
		args    args
		want    *moqtmessage.GroupMessage
		want1   io.Reader
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := tt.s.ReceiveDatagram(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Subscriber.ReceiveDatagram() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Subscriber.ReceiveDatagram() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Subscriber.ReceiveDatagram() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestSubscriber_SetupMOWT(t *testing.T) {
	type args struct {
		wtconn *webtransport.Session
	}
	tests := []struct {
		name    string
		s       Subscriber
		args    args
		want    *Session
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.SetupMOWT(tt.args.wtconn)
			if (err != nil) != tt.wantErr {
				t.Errorf("Subscriber.SetupMOWT() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Subscriber.SetupMOWT() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSubscriber_setupMOWT(t *testing.T) {
	type args struct {
		conn Connection
	}
	tests := []struct {
		name    string
		s       Subscriber
		args    args
		want    *Session
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.setupMOWT(tt.args.conn)
			if (err != nil) != tt.wantErr {
				t.Errorf("Subscriber.setupMOWT() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Subscriber.setupMOWT() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSubscriber_AcceptDataStream(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		s       Subscriber
		args    args
		want    *moqtmessage.GroupMessage
		want1   ReceiveStream
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := tt.s.AcceptDataStream(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Subscriber.AcceptDataStream() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Subscriber.AcceptDataStream() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Subscriber.AcceptDataStream() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
