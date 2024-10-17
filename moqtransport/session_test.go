package moqtransport

import (
	"context"
	"reflect"
	"testing"
)

func TestSession_OpenAnnounceStream(t *testing.T) {
	type args struct {
		stream Stream
	}
	tests := []struct {
		name    string
		sess    Session
		args    args
		want    *ReceiveAnnounceStream
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.sess.OpenAnnounceStream(tt.args.stream)
			if (err != nil) != tt.wantErr {
				t.Errorf("Session.OpenAnnounceStream() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Session.OpenAnnounceStream() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSession_AcceptAnnounceStream(t *testing.T) {
	type args struct {
		stream Stream
		ctx    context.Context
	}
	tests := []struct {
		name    string
		sess    Session
		args    args
		want    *SendAnnounceStream
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.sess.AcceptAnnounceStream(tt.args.stream, tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Session.AcceptAnnounceStream() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Session.AcceptAnnounceStream() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSession_OpenSubscribeStream(t *testing.T) {
	type args struct {
		stream Stream
	}
	tests := []struct {
		name    string
		sess    Session
		args    args
		want    *SendSubscribeStream
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.sess.OpenSubscribeStream(tt.args.stream)
			if (err != nil) != tt.wantErr {
				t.Errorf("Session.OpenSubscribeStream() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Session.OpenSubscribeStream() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSession_AcceptSubscribeStream(t *testing.T) {
	type args struct {
		stream Stream
		ctx    context.Context
	}
	tests := []struct {
		name    string
		sess    Session
		args    args
		want    *ReceiveSubscribeStream
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.sess.AcceptSubscribeStream(tt.args.stream, tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Session.AcceptSubscribeStream() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Session.AcceptSubscribeStream() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSession_PeekStreamType(t *testing.T) {
	type args struct {
		stream Stream
	}
	tests := []struct {
		name    string
		sess    Session
		args    args
		want    StreamType
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.sess.PeekStreamType(tt.args.stream)
			if (err != nil) != tt.wantErr {
				t.Errorf("Session.PeekStreamType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Session.PeekStreamType() = %v, want %v", got, tt.want)
			}
		})
	}
}
