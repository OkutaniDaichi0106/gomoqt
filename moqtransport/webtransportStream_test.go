package moqtransport

import (
	"reflect"
	"testing"
	"time"
)

func Test_webtransportStream_StreamID(t *testing.T) {
	tests := []struct {
		name    string
		wrapper webtransportStream
		want    StreamID
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.wrapper.StreamID(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("webtransportStream.StreamID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_webtransportStream_Read(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		wrapper webtransportStream
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.wrapper.Read(tt.args.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("webtransportStream.Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("webtransportStream.Read() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_webtransportStream_Write(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		wrapper webtransportStream
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.wrapper.Write(tt.args.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("webtransportStream.Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("webtransportStream.Write() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_webtransportStream_CancelRead(t *testing.T) {
	type args struct {
		code StreamErrorCode
	}
	tests := []struct {
		name    string
		wrapper webtransportStream
		args    args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wrapper.CancelRead(tt.args.code)
		})
	}
}

func Test_webtransportStream_CancelWrite(t *testing.T) {
	type args struct {
		code StreamErrorCode
	}
	tests := []struct {
		name    string
		wrapper webtransportStream
		args    args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wrapper.CancelWrite(tt.args.code)
		})
	}
}

func Test_webtransportStream_SetDeadLine(t *testing.T) {
	type args struct {
		time time.Time
	}
	tests := []struct {
		name    string
		wrapper webtransportStream
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.wrapper.SetDeadLine(tt.args.time); (err != nil) != tt.wantErr {
				t.Errorf("webtransportStream.SetDeadLine() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_webtransportStream_SetReadDeadline(t *testing.T) {
	type args struct {
		time time.Time
	}
	tests := []struct {
		name    string
		wrapper webtransportStream
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.wrapper.SetReadDeadline(tt.args.time); (err != nil) != tt.wantErr {
				t.Errorf("webtransportStream.SetReadDeadline() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_webtransportStream_SetWriteDeadline(t *testing.T) {
	type args struct {
		time time.Time
	}
	tests := []struct {
		name    string
		wrapper webtransportStream
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.wrapper.SetWriteDeadline(tt.args.time); (err != nil) != tt.wantErr {
				t.Errorf("webtransportStream.SetWriteDeadline() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_webtransportStream_Close(t *testing.T) {
	tests := []struct {
		name    string
		wrapper webtransportStream
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.wrapper.Close(); (err != nil) != tt.wantErr {
				t.Errorf("webtransportStream.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_webtransportStream_SetType(t *testing.T) {
	type args struct {
		streamType StreamType
	}
	tests := []struct {
		name    string
		wrapper *webtransportStream
		args    args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wrapper.SetType(tt.args.streamType)
		})
	}
}

func Test_webtransportStream_Type(t *testing.T) {
	tests := []struct {
		name    string
		wrapper webtransportStream
		want    StreamType
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.wrapper.Type(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("webtransportStream.Type() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_webtransportReceiveStream_StreamID(t *testing.T) {
	tests := []struct {
		name    string
		wrapper webtransportReceiveStream
		want    StreamID
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.wrapper.StreamID(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("webtransportReceiveStream.StreamID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_webtransportReceiveStream_Read(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		wrapper webtransportReceiveStream
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.wrapper.Read(tt.args.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("webtransportReceiveStream.Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("webtransportReceiveStream.Read() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_webtransportReceiveStream_CancelRead(t *testing.T) {
	type args struct {
		code StreamErrorCode
	}
	tests := []struct {
		name    string
		wrapper webtransportReceiveStream
		args    args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wrapper.CancelRead(tt.args.code)
		})
	}
}

func Test_webtransportReceiveStream_SetReadDeadline(t *testing.T) {
	type args struct {
		time time.Time
	}
	tests := []struct {
		name    string
		wrapper webtransportReceiveStream
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.wrapper.SetReadDeadline(tt.args.time); (err != nil) != tt.wantErr {
				t.Errorf("webtransportReceiveStream.SetReadDeadline() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_webtransportReceiveStream_Type(t *testing.T) {
	tests := []struct {
		name    string
		wrapper webtransportReceiveStream
		want    StreamType
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.wrapper.Type(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("webtransportReceiveStream.Type() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_webtransportSendStream_StreamID(t *testing.T) {
	tests := []struct {
		name    string
		wrapper webtransportSendStream
		want    StreamID
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.wrapper.StreamID(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("webtransportSendStream.StreamID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_webtransportSendStream_Write(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		wrapper webtransportSendStream
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.wrapper.Write(tt.args.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("webtransportSendStream.Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("webtransportSendStream.Write() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_webtransportSendStream_CancelWrite(t *testing.T) {
	type args struct {
		code StreamErrorCode
	}
	tests := []struct {
		name    string
		wrapper webtransportSendStream
		args    args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wrapper.CancelWrite(tt.args.code)
		})
	}
}

func Test_webtransportSendStream_SetWriteDeadline(t *testing.T) {
	type args struct {
		time time.Time
	}
	tests := []struct {
		name    string
		wrapper webtransportSendStream
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.wrapper.SetWriteDeadline(tt.args.time); (err != nil) != tt.wantErr {
				t.Errorf("webtransportSendStream.SetWriteDeadline() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_webtransportSendStream_Close(t *testing.T) {
	tests := []struct {
		name    string
		wrapper webtransportSendStream
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.wrapper.Close(); (err != nil) != tt.wantErr {
				t.Errorf("webtransportSendStream.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_webtransportSendStream_Type(t *testing.T) {
	tests := []struct {
		name    string
		wrapper webtransportSendStream
		want    StreamType
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.wrapper.Type(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("webtransportSendStream.Type() = %v, want %v", got, tt.want)
			}
		})
	}
}
