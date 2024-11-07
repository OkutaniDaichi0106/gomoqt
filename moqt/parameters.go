package moqt

import (
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/message"
)

type Parameters message.Parameters

func (p Parameters) Add(key uint64, value any) {
	message.Parameters(p).Add(key, value)
}

func (p Parameters) Remove(key uint64) {
	message.Parameters(p).Remove(key)
}

func (p Parameters) ReadAsByteArray(key uint64) ([]byte, bool) {
	return message.Parameters(p).AsByteArray(key)
}

func (p Parameters) ReadAsString(key uint64) (string, bool) {
	return message.Parameters(p).AsString(key)
}

func (p Parameters) ReadAsInt(key uint64) (int64, bool) {
	return message.Parameters(p).AsInt(key)
}

func (p Parameters) ReadAsUint(key uint64) (uint64, bool) {
	return message.Parameters(p).AsUint(key)
}

func (p Parameters) ReadAsBool(key uint64) (bool, bool) {
	return message.Parameters(p).AsBool(key)
}

const (
	ROLE               uint64 = 0x00
	PATH               uint64 = 0x01
	MAX_SUBSCRIBE_ID   uint64 = 0x02
	AUTHORIZATION_INFO uint64 = 0x03
	DELIVERY_TIMEOUT   uint64 = 0x04
	MAX_CACHE_DURATION uint64 = 0x05
)

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
