package moqtransport

import (
	"context"
	"net"
	"reflect"
	"testing"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/webtransport-go"
)

func Test_newMOWTConnection(t *testing.T) {
	type args struct {
		wtconn *webtransport.Session
	}
	tests := []struct {
		name string
		args args
		want Connection
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newMOWTConnection(tt.args.wtconn); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newMOWTConnection() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_webtransportConnection_AcceptStream(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		conn    *webtransportConnection
		args    args
		want    Stream
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.conn.AcceptStream(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("webtransportConnection.AcceptStream() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("webtransportConnection.AcceptStream() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_webtransportConnection_AcceptUniStream(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		conn    *webtransportConnection
		args    args
		want    ReceiveStream
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.conn.AcceptUniStream(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("webtransportConnection.AcceptUniStream() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("webtransportConnection.AcceptUniStream() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_webtransportConnection_CloseWithError(t *testing.T) {
	type args struct {
		code SessionErrorCode
		msg  string
	}
	tests := []struct {
		name    string
		conn    *webtransportConnection
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.conn.CloseWithError(tt.args.code, tt.args.msg); (err != nil) != tt.wantErr {
				t.Errorf("webtransportConnection.CloseWithError() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_webtransportConnection_ConnectionState(t *testing.T) {
	tests := []struct {
		name string
		conn *webtransportConnection
		want quic.ConnectionState
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.conn.ConnectionState(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("webtransportConnection.ConnectionState() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_webtransportConnection_Context(t *testing.T) {
	tests := []struct {
		name string
		conn *webtransportConnection
		want context.Context
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.conn.Context(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("webtransportConnection.Context() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_webtransportConnection_LocalAddr(t *testing.T) {
	tests := []struct {
		name string
		conn *webtransportConnection
		want net.Addr
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.conn.LocalAddr(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("webtransportConnection.LocalAddr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_webtransportConnection_OpenStream(t *testing.T) {
	tests := []struct {
		name    string
		conn    *webtransportConnection
		want    Stream
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.conn.OpenStream()
			if (err != nil) != tt.wantErr {
				t.Errorf("webtransportConnection.OpenStream() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("webtransportConnection.OpenStream() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_webtransportConnection_OpenStreamSync(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		conn    *webtransportConnection
		args    args
		want    Stream
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.conn.OpenStreamSync(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("webtransportConnection.OpenStreamSync() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("webtransportConnection.OpenStreamSync() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_webtransportConnection_OpenUniStream(t *testing.T) {
	tests := []struct {
		name    string
		conn    *webtransportConnection
		want    SendStream
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.conn.OpenUniStream()
			if (err != nil) != tt.wantErr {
				t.Errorf("webtransportConnection.OpenUniStream() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("webtransportConnection.OpenUniStream() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_webtransportConnection_OpenUniStreamSync(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		conn    *webtransportConnection
		args    args
		want    SendStream
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.conn.OpenUniStreamSync(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("webtransportConnection.OpenUniStreamSync() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("webtransportConnection.OpenUniStreamSync() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_webtransportConnection_ReceiveDatagram(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		conn    *webtransportConnection
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.conn.ReceiveDatagram(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("webtransportConnection.ReceiveDatagram() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("webtransportConnection.ReceiveDatagram() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_webtransportConnection_RemoteAddr(t *testing.T) {
	tests := []struct {
		name string
		conn *webtransportConnection
		want net.Addr
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.conn.RemoteAddr(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("webtransportConnection.RemoteAddr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_webtransportConnection_SendDatagram(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		conn    *webtransportConnection
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.conn.SendDatagram(tt.args.b); (err != nil) != tt.wantErr {
				t.Errorf("webtransportConnection.SendDatagram() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
