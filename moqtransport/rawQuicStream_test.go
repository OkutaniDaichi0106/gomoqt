package moqtransport

import (
	"reflect"
	"testing"
	"time"
)

func Test_rawQuicStream_StreamID(t *testing.T) {
	tests := []struct {
		name    string
		wrapper rawQuicStream
		want    StreamID
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.wrapper.StreamID(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("rawQuicStream.StreamID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_rawQuicStream_Read(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		wrapper rawQuicStream
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
				t.Errorf("rawQuicStream.Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("rawQuicStream.Read() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_rawQuicStream_Write(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		wrapper rawQuicStream
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
				t.Errorf("rawQuicStream.Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("rawQuicStream.Write() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_rawQuicStream_CancelRead(t *testing.T) {
	type args struct {
		code StreamErrorCode
	}
	tests := []struct {
		name    string
		wrapper rawQuicStream
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

func Test_rawQuicStream_CancelWrite(t *testing.T) {
	type args struct {
		code StreamErrorCode
	}
	tests := []struct {
		name    string
		wrapper rawQuicStream
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

func Test_rawQuicStream_SetDeadLine(t *testing.T) {
	type args struct {
		time time.Time
	}
	tests := []struct {
		name    string
		wrapper rawQuicStream
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.wrapper.SetDeadLine(tt.args.time); (err != nil) != tt.wantErr {
				t.Errorf("rawQuicStream.SetDeadLine() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_rawQuicStream_SetReadDeadline(t *testing.T) {
	type args struct {
		time time.Time
	}
	tests := []struct {
		name    string
		wrapper rawQuicStream
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.wrapper.SetReadDeadline(tt.args.time); (err != nil) != tt.wantErr {
				t.Errorf("rawQuicStream.SetReadDeadline() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_rawQuicStream_SetWriteDeadline(t *testing.T) {
	type args struct {
		time time.Time
	}
	tests := []struct {
		name    string
		wrapper rawQuicStream
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.wrapper.SetWriteDeadline(tt.args.time); (err != nil) != tt.wantErr {
				t.Errorf("rawQuicStream.SetWriteDeadline() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_rawQuicStream_Close(t *testing.T) {
	tests := []struct {
		name    string
		wrapper rawQuicStream
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.wrapper.Close(); (err != nil) != tt.wantErr {
				t.Errorf("rawQuicStream.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_rawQuicStream_SetType(t *testing.T) {
	type args struct {
		streamType StreamType
	}
	tests := []struct {
		name    string
		wrapper *rawQuicStream
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

func Test_rawQuicStream_Type(t *testing.T) {
	tests := []struct {
		name    string
		wrapper rawQuicStream
		want    StreamType
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.wrapper.Type(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("rawQuicStream.Type() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_rawQuicReceiveStream_StreamID(t *testing.T) {
	tests := []struct {
		name    string
		wrapper rawQuicReceiveStream
		want    StreamID
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.wrapper.StreamID(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("rawQuicReceiveStream.StreamID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_rawQuicReceiveStream_Read(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		wrapper rawQuicReceiveStream
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
				t.Errorf("rawQuicReceiveStream.Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("rawQuicReceiveStream.Read() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_rawQuicReceiveStream_CancelRead(t *testing.T) {
	type args struct {
		code StreamErrorCode
	}
	tests := []struct {
		name    string
		wrapper rawQuicReceiveStream
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

func Test_rawQuicReceiveStream_SetReadDeadline(t *testing.T) {
	type args struct {
		time time.Time
	}
	tests := []struct {
		name    string
		wrapper rawQuicReceiveStream
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.wrapper.SetReadDeadline(tt.args.time); (err != nil) != tt.wantErr {
				t.Errorf("rawQuicReceiveStream.SetReadDeadline() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_rawQuicReceiveStream_Type(t *testing.T) {
	tests := []struct {
		name    string
		wrapper rawQuicReceiveStream
		want    StreamType
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.wrapper.Type(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("rawQuicReceiveStream.Type() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_rawQuicSendStream_StreamID(t *testing.T) {
	tests := []struct {
		name    string
		wrapper rawQuicSendStream
		want    StreamID
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.wrapper.StreamID(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("rawQuicSendStream.StreamID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_rawQuicSendStream_Write(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		wrapper rawQuicSendStream
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
				t.Errorf("rawQuicSendStream.Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("rawQuicSendStream.Write() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_rawQuicSendStream_CancelWrite(t *testing.T) {
	type args struct {
		code StreamErrorCode
	}
	tests := []struct {
		name    string
		wrapper rawQuicSendStream
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

func Test_rawQuicSendStream_SetWriteDeadline(t *testing.T) {
	type args struct {
		time time.Time
	}
	tests := []struct {
		name    string
		wrapper rawQuicSendStream
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.wrapper.SetWriteDeadline(tt.args.time); (err != nil) != tt.wantErr {
				t.Errorf("rawQuicSendStream.SetWriteDeadline() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_rawQuicSendStream_Close(t *testing.T) {
	tests := []struct {
		name    string
		wrapper rawQuicSendStream
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.wrapper.Close(); (err != nil) != tt.wantErr {
				t.Errorf("rawQuicSendStream.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_rawQuicSendStream_Type(t *testing.T) {
	tests := []struct {
		name    string
		wrapper rawQuicSendStream
		want    StreamType
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.wrapper.Type(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("rawQuicSendStream.Type() = %v, want %v", got, tt.want)
			}
		})
	}
}
