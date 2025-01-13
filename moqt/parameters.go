package moqt

import (
	"bytes"
	"errors"
	"log/slog"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/quic-go/quic-go/quicvarint"
)

type Parameters message.Parameters

func NewParameters() Parameters {
	return make(Parameters)
}

func (p Parameters) SetByteArray(key uint64, value []byte) {
	p[key] = value
}

func (p Parameters) SetString(key uint64, value string) {
	p[key] = []byte(value)
}

func (p Parameters) SetInt(key uint64, value int64) {
	p[key] = quicvarint.Append(make([]byte, 0), uint64(value))
}

func (p Parameters) SetUint(key uint64, value uint64) {
	p[key] = quicvarint.Append(make([]byte, 0), value)
}

func (p Parameters) SetBool(key uint64, value bool) {
	if value {
		p[key] = quicvarint.Append(make([]byte, 0), 1)
	} else {
		p[key] = quicvarint.Append(make([]byte, 0), 0)
	}
}

func (p Parameters) Remove(key uint64) {
	delete(p, key)
}

func (p Parameters) GetByteArray(key uint64) ([]byte, error) {
	value, ok := p[key]
	if !ok {
		return nil, ErrParameterNotFound
	}

	return value, nil
}

func (p Parameters) GetString(key uint64) (string, error) {
	value, err := p.GetByteArray(key)
	if err != nil {
		slog.Error("failed to read a parameter as byte array")
		return "", err
	}

	return string(value), nil
}

func (p Parameters) GetInt(key uint64) (int64, error) {
	num, err := p.GetUint(key)
	if err != nil {
		slog.Error("failed to read a parameter as uint", slog.String("error", err.Error()))
		return 0, err
	}

	return int64(num), nil
}

func (p Parameters) GetUint(key uint64) (uint64, error) {
	value, ok := p[key]
	if !ok {
		return 0, ErrParameterNotFound
	}

	num, err := quicvarint.Read(quicvarint.NewReader(bytes.NewReader(value)))
	if err != nil {
		slog.Error("failed to read the bytes as uint64")
		return 0, err
	}

	return num, nil
}

func (p Parameters) GetBool(key uint64) (bool, error) {
	num, err := p.GetUint(key)
	if err != nil {
		slog.Error("failed to read a parameter as uint", slog.String("error", err.Error()))
		return false, err
	}

	switch num {
	case 0:
		return false, nil
	case 1:
		return true, nil
	default:
		return false, errors.New("invalid value as bool")
	}
}

var ErrParameterNotFound = errors.New("parameter not found")

/***/
const (
	path               uint64 = 0x01
	max_subscribe_id   uint64 = 0x02
	authorization_info uint64 = 0x03
	delivery_timeout   uint64 = 0x04
	max_cache_duration uint64 = 0x05
)

// Path parameter
func (p Parameters) SetPath(value string) {
	p.SetString(path, value)
}

func (p Parameters) GetPath() (string, bool) {
	num, err := p.GetString(path)
	if err != nil {
		return "", false
	}
	return num, true
}

// MaxSubscribeID parameter
func (p Parameters) SetMaxSubscribeID(value SubscribeID) {
	p.SetUint(max_subscribe_id, uint64(value))
}

func (p Parameters) GetMaxSubscribeID() (SubscribeID, bool) {
	num, err := p.GetUint(max_subscribe_id)
	if err != nil {
		return 0, false
	}

	return SubscribeID(num), true
}

// MaxCacheDuration parameter
func (p Parameters) SetMaxCacheDuration(value time.Duration) {
	p.SetInt(max_cache_duration, int64(value))
}

func (p Parameters) GetMaxCacheDuration() (time.Duration, bool) {
	num, err := p.GetUint(max_cache_duration)
	if err != nil {
		return 0, false
	}

	return time.Duration(num), true
}

// AuthorizationInfo parameter
func (p Parameters) SetAuthorizationInfo(value string) {
	p.SetString(authorization_info, value)
}

func (p Parameters) GetAuthorizationInfo() (string, bool) {
	str, err := p.GetString(authorization_info)
	if err != nil {
		return "", false
	}

	return str, true
}

// DeliveryTimeout parameter
func (p Parameters) SetDeliveryTimeout(value time.Duration) {
	p.SetInt(delivery_timeout, int64(value))
}

func (p Parameters) GetDeliveryTimeout() (time.Duration, bool) {
	num, err := p.GetUint(delivery_timeout)
	if err != nil {
		return 0, false
	}

	return time.Duration(num), true
}
