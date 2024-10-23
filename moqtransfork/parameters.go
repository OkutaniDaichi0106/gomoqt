package moqtransfork

import (
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqtransfork/internal/message"
)

type ParameterKey uint64

const (
	ROLE               uint64 = 0x00
	PATH               uint64 = 0x01
	MAX_SUBSCRIBE_ID   uint64 = 0x02
	AUTHORIZATION_INFO uint64 = 0x03
	DELIVERY_TIMEOUT   uint64 = 0x04
	MAX_CACHE_DURATION uint64 = 0x05
)

type Parameters struct {
	message.Parameters
}

type Role message.Role
type SubscribeID message.SubscribeID

func (params Parameters) Role() (Role, bool) {
	num, ok := params.AsUint(ROLE)
	if !ok {
		return 0, false
	}
	return Role(num), true
}

func (params Parameters) Path() (string, bool) {
	num, ok := params.AsString(PATH)
	if !ok {
		return "", false
	}
	return num, true
}

func (params Parameters) MaxSubscribeID() (SubscribeID, bool) {
	num, ok := params.AsUint(MAX_SUBSCRIBE_ID)
	if !ok {
		return 0, false
	}

	return SubscribeID(num), true
}

func (params Parameters) MaxCacheDuration() (time.Duration, bool) {
	num, ok := params.AsUint(MAX_CACHE_DURATION)
	if !ok {
		return 0, false
	}

	return time.Duration(num), true
}

func (params Parameters) AuthorizationInfo() (string, bool) {
	str, ok := params.AsString(AUTHORIZATION_INFO)
	if !ok {
		return "", false
	}

	return str, true
}

func (params Parameters) DeliveryTimeout() (time.Duration, bool) {
	num, ok := params.AsUint(DELIVERY_TIMEOUT)
	if !ok {
		return 0, false
	}

	return time.Duration(num), true
}
