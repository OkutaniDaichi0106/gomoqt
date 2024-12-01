package moqt

import (
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
)

type Parameters message.Parameters

func (p Parameters) Add(key uint64, value any) {
	switch v := value.(type) {
	case int64:
		p[key] = uint64(v)
	case int32:
		p[key] = uint64(v)
	case int16:
		p[key] = uint64(v)
	case int8:
		p[key] = uint64(v)
	case uint32:
		p[key] = uint64(v)
	case uint16:
		p[key] = uint64(v)
	case uint8:
		p[key] = uint64(v)
	case uint64:
		p[key] = v
	case bool:
		if v {
			p[key] = 1
		} else if !v {
			p[key] = 0
		}
	case string:
		p[key] = []byte(v)
	case []byte:
		p[key] = v
	default:
		panic("invalid type")
	}
}

func (p Parameters) Remove(key uint64) {
	delete(p, key)
}

func (p Parameters) ReadAsByteArray(key uint64) ([]byte, bool) {
	value, ok := p[key]
	if !ok {
		return nil, false
	}
	switch v := value.(type) {
	case []byte:
		return v, true
	default:
		return nil, false
	}
}

func (p Parameters) ReadAsString(key uint64) (string, bool) {
	value, ok := p.ReadAsByteArray(key)
	if !ok {
		return "", false
	}
	return string(value), true
}

func (p Parameters) ReadAsInt(key uint64) (int64, bool) {
	value, ok := p[key]
	if !ok {
		return 0, false
	}
	switch v := value.(type) {
	case int:
		return int64(v), true
	case int8:
		return int64(v), true
	case int16:
		return int64(v), true
	case int32:
		return int64(v), true
	case int64:
		return v, true
	case uint:
		return int64(v), true
	case uint8:
		return int64(v), true
	case uint16:
		return int64(v), true
	case uint32:
		return int64(v), true
	case uint64:
		return int64(v), true
	default:
		return 0, false
	}
}

func (p Parameters) ReadAsUint(key uint64) (uint64, bool) {
	value, ok := p[key]
	if !ok {
		return 0, false
	}
	switch v := value.(type) {
	case int:
		return uint64(v), true
	case int8:
		return uint64(v), true
	case int16:
		return uint64(v), true
	case int32:
		return uint64(v), true
	case int64:
		return uint64(v), true
	case uint:
		return uint64(v), true
	case uint8:
		return uint64(v), true
	case uint16:
		return uint64(v), true
	case uint32:
		return uint64(v), true
	case uint64:
		return v, true
	default:
		return 0, false
	}
}

func (p Parameters) ReadAsBool(key uint64) (bool, bool) {
	value, ok := p[key]
	if !ok {
		return false, false
	}
	switch v := value.(type) {
	case uint64:
		if v == 0 {
			return false, true
		} else if v == 1 {
			return true, true
		} else {
			return false, false
		}
	default:
		return false, false
	}
}

const (
	ROLE               uint64 = 0x00
	PATH               uint64 = 0x01
	MAX_SUBSCRIBE_ID   uint64 = 0x02
	AUTHORIZATION_INFO uint64 = 0x03
	DELIVERY_TIMEOUT   uint64 = 0x04
	MAX_CACHE_DURATION uint64 = 0x05
)

func getPath(params Parameters) (string, bool) {
	num, ok := params.ReadAsString(PATH)
	if !ok {
		return "", false
	}
	return num, true
}

func getMaxSubscribeID(params Parameters) (SubscribeID, bool) {
	num, ok := params.ReadAsUint(MAX_SUBSCRIBE_ID)
	if !ok {
		return 0, false
	}

	return SubscribeID(num), true
}

func getMaxCacheDuration(params Parameters) (time.Duration, bool) {
	num, ok := params.ReadAsUint(MAX_CACHE_DURATION)
	if !ok {
		return 0, false
	}

	return time.Duration(num), true
}

func getAuthorizationInfo(params Parameters) (string, bool) {
	str, ok := params.ReadAsString(AUTHORIZATION_INFO)
	if !ok {
		return "", false
	}

	return str, true
}

func getDeliveryTimeout(params Parameters) (time.Duration, bool) {
	num, ok := params.ReadAsUint(DELIVERY_TIMEOUT)
	if !ok {
		return 0, false
	}

	return time.Duration(num), true
}
