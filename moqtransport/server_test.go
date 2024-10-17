package moqtransport

import (
	"crypto/tls"
	"reflect"
	"testing"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/webtransport-go"
)

func TestServer_ListenAndServeQUIC(t *testing.T) {
	type args struct {
		addr       string
		handler    QUICHandler
		tlsConfig  *tls.Config
		quicConfig *quic.Config
	}
	tests := []struct {
		name    string
		s       Server
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.ListenAndServeQUIC(tt.args.addr, tt.args.handler, tt.args.tlsConfig, tt.args.quicConfig); (err != nil) != tt.wantErr {
				t.Errorf("Server.ListenAndServeQUIC() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestServer_ListenAndServeWT(t *testing.T) {
	type args struct {
		wts *webtransport.Server
	}
	tests := []struct {
		name    string
		s       Server
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.ListenAndServeWT(tt.args.wts); (err != nil) != tt.wantErr {
				t.Errorf("Server.ListenAndServeWT() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestServer_SetupMORQ(t *testing.T) {
	type args struct {
		qconn quic.Connection
	}
	tests := []struct {
		name    string
		s       Server
		args    args
		want    *Session
		want1   string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := tt.s.SetupMORQ(tt.args.qconn)
			if (err != nil) != tt.wantErr {
				t.Errorf("Server.SetupMORQ() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Server.SetupMORQ() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Server.SetupMORQ() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestServer_setupMORQ(t *testing.T) {
	type args struct {
		conn Connection
	}
	tests := []struct {
		name    string
		s       Server
		args    args
		want    *Session
		want1   string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := tt.s.setupMORQ(tt.args.conn)
			if (err != nil) != tt.wantErr {
				t.Errorf("Server.setupMORQ() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Server.setupMORQ() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Server.setupMORQ() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestServer_SetupMOWT(t *testing.T) {
	type args struct {
		wtconn *webtransport.Session
	}
	tests := []struct {
		name    string
		s       Server
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
				t.Errorf("Server.SetupMOWT() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Server.SetupMOWT() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServer_setupMOWT(t *testing.T) {
	type args struct {
		conn Connection
	}
	tests := []struct {
		name    string
		s       Server
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
				t.Errorf("Server.setupMOWT() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Server.setupMOWT() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_acceptSetupStream(t *testing.T) {
	type args struct {
		stream Stream
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := acceptSetupStream(tt.args.stream); (err != nil) != tt.wantErr {
				t.Errorf("acceptSetupStream() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestServer_SetCertFiles(t *testing.T) {
	type args struct {
		certFile string
		keyFile  string
	}
	tests := []struct {
		name    string
		s       *Server
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.SetCertFiles(tt.args.certFile, tt.args.keyFile); (err != nil) != tt.wantErr {
				t.Errorf("Server.SetCertFiles() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
