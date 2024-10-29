package moqt

import (
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
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

type Parameters message.Parameters

type Role byte

type SubscribeID uint64

func getRole(params message.Parameters) (Role, bool) {
	num, ok := params.AsUint(ROLE)
	if !ok {
		return 0, false
	}
	return Role(num), true
}

func getPath(params message.Parameters) (string, bool) {
	num, ok := params.AsString(PATH)
	if !ok {
		return "", false
	}
	return num, true
}

func getMaxSubscribeID(params message.Parameters) (SubscribeID, bool) {
	num, ok := params.AsUint(MAX_SUBSCRIBE_ID)
	if !ok {
		return 0, false
	}

	return SubscribeID(num), true
}

func getMaxCacheDuration(params message.Parameters) (time.Duration, bool) {
	num, ok := params.AsUint(MAX_CACHE_DURATION)
	if !ok {
		return 0, false
	}

	return time.Duration(num), true
}

func getAuthorizationInfo(params message.Parameters) (string, bool) {
	str, ok := params.AsString(AUTHORIZATION_INFO)
	if !ok {
		return "", false
	}

	return str, true
}

func getDeliveryTimeout(params message.Parameters) (time.Duration, bool) {
	num, ok := params.AsUint(DELIVERY_TIMEOUT)
	if !ok {
		return 0, false
	}

	return time.Duration(num), true
}
