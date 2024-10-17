package moqtransport

import (
	"reflect"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/internal/moqtmessage"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/webtransport-go"
)

func TestPublisher_SetupMORQ(t *testing.T) {
	type args struct {
		qconn quic.Connection
		path  string
	}
	tests := []struct {
		name    string
		p       *Publisher
		args    args
		want    *Session
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.p.SetupMORQ(tt.args.qconn, tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Publisher.SetupMORQ() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Publisher.SetupMORQ() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPublisher_setupMORQ(t *testing.T) {
	type args struct {
		conn Connection
		path string
	}
	tests := []struct {
		name    string
		p       *Publisher
		args    args
		want    *Session
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.p.setupMORQ(tt.args.conn, tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Publisher.setupMORQ() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Publisher.setupMORQ() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPublisher_SetupMOWT(t *testing.T) {
	type args struct {
		wtconn *webtransport.Session
	}
	tests := []struct {
		name    string
		p       *Publisher
		args    args
		want    *Session
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.p.SetupMOWT(tt.args.wtconn)
			if (err != nil) != tt.wantErr {
				t.Errorf("Publisher.SetupMOWT() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Publisher.SetupMOWT() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPublisher_setupMOWT(t *testing.T) {
	type args struct {
		conn Connection
	}
	tests := []struct {
		name    string
		p       *Publisher
		args    args
		want    *Session
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.p.setupMOWT(tt.args.conn)
			if (err != nil) != tt.wantErr {
				t.Errorf("Publisher.setupMOWT() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Publisher.setupMOWT() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPublisher_SendDatagram(t *testing.T) {
	type args struct {
		group   moqtmessage.GroupMessage
		payload []byte
	}
	tests := []struct {
		name    string
		p       Publisher
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.p.SendDatagram(tt.args.group, tt.args.payload); (err != nil) != tt.wantErr {
				t.Errorf("Publisher.SendDatagram() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPublisher_OpenDataStream(t *testing.T) {
	type args struct {
		group moqtmessage.GroupMessage
	}
	tests := []struct {
		name    string
		p       Publisher
		args    args
		want    SendStream
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.p.OpenDataStream(tt.args.group)
			if (err != nil) != tt.wantErr {
				t.Errorf("Publisher.OpenDataStream() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Publisher.OpenDataStream() = %v, want %v", got, tt.want)
			}
		})
	}
}
