package moqtransport

import (
	"context"
	"net"
	"reflect"
	"testing"

	"github.com/quic-go/quic-go"
)

func Test_newMORQConnection(t *testing.T) {
	type args struct {
		conn quic.Connection
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
			if got := newMORQConnection(tt.args.conn); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newMORQConnection() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_rawQuicConnection_AcceptStream(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		wrapper *rawQuicConnection
		args    args
		want    Stream
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.wrapper.AcceptStream(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("rawQuicConnection.AcceptStream() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("rawQuicConnection.AcceptStream() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_rawQuicConnection_AcceptUniStream(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		wrapper *rawQuicConnection
		args    args
		want    ReceiveStream
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.wrapper.AcceptUniStream(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("rawQuicConnection.AcceptUniStream() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("rawQuicConnection.AcceptUniStream() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_rawQuicConnection_CloseWithError(t *testing.T) {
	type args struct {
		code SessionErrorCode
		msg  string
	}
	tests := []struct {
		name    string
		wrapper *rawQuicConnection
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.wrapper.CloseWithError(tt.args.code, tt.args.msg); (err != nil) != tt.wantErr {
				t.Errorf("rawQuicConnection.CloseWithError() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_rawQuicConnection_ConnectionState(t *testing.T) {
	tests := []struct {
		name    string
		wrapper *rawQuicConnection
		want    quic.ConnectionState
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.wrapper.ConnectionState(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("rawQuicConnection.ConnectionState() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_rawQuicConnection_Context(t *testing.T) {
	tests := []struct {
		name    string
		wrapper *rawQuicConnection
		want    context.Context
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.wrapper.Context(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("rawQuicConnection.Context() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_rawQuicConnection_LocalAddr(t *testing.T) {
	tests := []struct {
		name    string
		wrapper *rawQuicConnection
		want    net.Addr
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.wrapper.LocalAddr(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("rawQuicConnection.LocalAddr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_rawQuicConnection_OpenStream(t *testing.T) {
	tests := []struct {
		name    string
		wrapper *rawQuicConnection
		want    Stream
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.wrapper.OpenStream()
			if (err != nil) != tt.wantErr {
				t.Errorf("rawQuicConnection.OpenStream() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("rawQuicConnection.OpenStream() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_rawQuicConnection_OpenStreamSync(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		wrapper *rawQuicConnection
		args    args
		want    Stream
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.wrapper.OpenStreamSync(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("rawQuicConnection.OpenStreamSync() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("rawQuicConnection.OpenStreamSync() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_rawQuicConnection_OpenUniStream(t *testing.T) {
	tests := []struct {
		name    string
		wrapper *rawQuicConnection
		want    SendStream
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.wrapper.OpenUniStream()
			if (err != nil) != tt.wantErr {
				t.Errorf("rawQuicConnection.OpenUniStream() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("rawQuicConnection.OpenUniStream() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_rawQuicConnection_OpenUniStreamSync(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		wrapper *rawQuicConnection
		args    args
		want    SendStream
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.wrapper.OpenUniStreamSync(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("rawQuicConnection.OpenUniStreamSync() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("rawQuicConnection.OpenUniStreamSync() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_rawQuicConnection_ReceiveDatagram(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		wrapper *rawQuicConnection
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.wrapper.ReceiveDatagram(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("rawQuicConnection.ReceiveDatagram() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("rawQuicConnection.ReceiveDatagram() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_rawQuicConnection_RemoteAddr(t *testing.T) {
	tests := []struct {
		name    string
		wrapper *rawQuicConnection
		want    net.Addr
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.wrapper.RemoteAddr(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("rawQuicConnection.RemoteAddr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_rawQuicConnection_SendDatagram(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		wrapper *rawQuicConnection
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.wrapper.SendDatagram(tt.args.b); (err != nil) != tt.wantErr {
				t.Errorf("rawQuicConnection.SendDatagram() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_isValidPath(t *testing.T) {
	type args struct {
		pattern string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidPath(tt.args.pattern); got != tt.want {
				t.Errorf("isValidPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
