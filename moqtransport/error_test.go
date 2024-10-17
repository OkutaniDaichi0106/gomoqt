package moqtransport

import (
	"reflect"
	"testing"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransport/internal/moqtmessage"
)

func TestDefaultAnnounceError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  DefaultAnnounceError
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("DefaultAnnounceError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultAnnounceError_Code(t *testing.T) {
	tests := []struct {
		name string
		err  DefaultAnnounceError
		want AnnounceErrorCode
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Code(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DefaultAnnounceError.Code() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInternalError_Error(t *testing.T) {
	tests := []struct {
		name string
		i    InternalError
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.i.Error(); got != tt.want {
				t.Errorf("InternalError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInternalError_AnnounceErrorCode(t *testing.T) {
	tests := []struct {
		name string
		i    InternalError
		want AnnounceErrorCode
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.i.AnnounceErrorCode(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InternalError.AnnounceErrorCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInternalError_SubscribeErrorCode(t *testing.T) {
	tests := []struct {
		name string
		i    InternalError
		want SubscribeErrorCode
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.i.SubscribeErrorCode(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InternalError.SubscribeErrorCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInternalError_SubscribeDoneErrorCode(t *testing.T) {
	tests := []struct {
		name string
		i    InternalError
		want SubscribeDoneStatusCode
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.i.SubscribeDoneErrorCode(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InternalError.SubscribeDoneErrorCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInternalError_TerminateErrorCode(t *testing.T) {
	tests := []struct {
		name string
		i    InternalError
		want TerminateErrorCode
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.i.TerminateErrorCode(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InternalError.TerminateErrorCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnauthorizedError_Error(t *testing.T) {
	tests := []struct {
		name string
		u    UnauthorizedError
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.u.Error(); got != tt.want {
				t.Errorf("UnauthorizedError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnauthorizedError_SubscribeErrorCode(t *testing.T) {
	tests := []struct {
		name string
		u    UnauthorizedError
		want SubscribeErrorCode
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.u.SubscribeErrorCode(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UnauthorizedError.SubscribeErrorCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnauthorizedError_SubscribeDoneErrorCode(t *testing.T) {
	tests := []struct {
		name string
		u    UnauthorizedError
		want SubscribeDoneStatusCode
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.u.SubscribeDoneErrorCode(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UnauthorizedError.SubscribeDoneErrorCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnauthorizedError_TerminateErrorCode(t *testing.T) {
	tests := []struct {
		name string
		u    UnauthorizedError
		want TerminateErrorCode
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.u.TerminateErrorCode(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UnauthorizedError.TerminateErrorCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultSubscribeError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  DefaultSubscribeError
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("DefaultSubscribeError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultSubscribeError_SubscribeErrorCode(t *testing.T) {
	tests := []struct {
		name string
		err  DefaultSubscribeError
		want SubscribeErrorCode
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.SubscribeErrorCode(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DefaultSubscribeError.SubscribeErrorCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRetryTrackAliasError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  RetryTrackAliasError
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("RetryTrackAliasError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRetryTrackAliasError_SubscribeErrorCode(t *testing.T) {
	tests := []struct {
		name string
		err  RetryTrackAliasError
		want SubscribeErrorCode
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.SubscribeErrorCode(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RetryTrackAliasError.SubscribeErrorCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRetryTrackAliasError_TrackAlias(t *testing.T) {
	tests := []struct {
		name string
		err  RetryTrackAliasError
		want moqtmessage.TrackAlias
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.TrackAlias(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RetryTrackAliasError.TrackAlias() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultSubscribeDoneError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  DefaultSubscribeDoneError
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("DefaultSubscribeDoneError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultSubscribeDoneError_Reason(t *testing.T) {
	tests := []struct {
		name string
		err  DefaultSubscribeDoneError
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Reason(); got != tt.want {
				t.Errorf("DefaultSubscribeDoneError.Reason() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultSubscribeDoneError_SubscribeDoneErrorCode(t *testing.T) {
	tests := []struct {
		name string
		err  DefaultSubscribeDoneError
		want SubscribeDoneStatusCode
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.SubscribeDoneErrorCode(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DefaultSubscribeDoneError.SubscribeDoneErrorCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultSubscribeDoneStatus_Reason(t *testing.T) {
	tests := []struct {
		name   string
		status DefaultSubscribeDoneStatus
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.Reason(); got != tt.want {
				t.Errorf("DefaultSubscribeDoneStatus.Reason() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultSubscribeDoneStatus_Code(t *testing.T) {
	tests := []struct {
		name   string
		status DefaultSubscribeDoneStatus
		want   SubscribeDoneStatusCode
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.Code(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DefaultSubscribeDoneStatus.Code() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultTerminateError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  DefaultTerminateError
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("DefaultTerminateError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultTerminateError_TerminateErrorCode(t *testing.T) {
	tests := []struct {
		name string
		err  DefaultTerminateError
		want TerminateErrorCode
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.TerminateErrorCode(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DefaultTerminateError.TerminateErrorCode() = %v, want %v", got, tt.want)
			}
		})
	}
}
