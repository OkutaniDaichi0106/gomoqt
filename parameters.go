package moqt

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/quic-go/quic-go/quicvarint"
)

type Parameters message.Parameters

func (p Parameters) Add(key uint64, value any) error {
	encode := func(v uint64) []byte {
		value := make([]byte, quicvarint.Len(uint64(v)))
		return quicvarint.Append(value, uint64(v))
	}
	switch v := value.(type) {
	case int64, int32, int16, int8:
		p[key] = encode(uint64(reflect.ValueOf(v).Int()))
	case uint64, uint32, uint16, uint8:
		p[key] = encode(reflect.ValueOf(v).Uint())
	case bool:
		if v {
			p[key] = encode(1)
		} else if !v {
			p[key] = encode(0)
		}
	case string:
		p[key] = []byte(v)
	case []byte:
		p[key] = v
	default:
		return fmt.Errorf("invalid type: %T", value)
	}

	return nil
}

func (p Parameters) Remove(key uint64) {
	delete(p, key)
}

func (p Parameters) ReadAsByteArray(key uint64) ([]byte, error) {
	value, ok := p[key]
	if !ok {
		return nil, ErrParameterNotFound
	}

	return value, nil
}

func (p Parameters) ReadAsString(key uint64) (string, error) {
	value, err := p.ReadAsByteArray(key)
	if err != nil {
		slog.Error("failed to read a parameter as byte array")
		return "", err
	}

	return string(value), nil
}

func (p Parameters) ReadAsInt(key uint64) (int64, error) {
	num, err := p.ReadAsUint(key)
	if err != nil {
		slog.Error("failed to read a parameter as uint", slog.String("error", err.Error()))
		return 0, err
	}

	return int64(num), nil
}

func (p Parameters) ReadAsUint(key uint64) (uint64, error) {
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

func (p Parameters) ReadAsBool(key uint64) (bool, error) {
	num, err := p.ReadAsUint(key)
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

const (
	// ROLE               uint64 = 0x00
	PATH               uint64 = 0x01
	MAX_SUBSCRIBE_ID   uint64 = 0x02
	AUTHORIZATION_INFO uint64 = 0x03
	DELIVERY_TIMEOUT   uint64 = 0x04
	MAX_CACHE_DURATION uint64 = 0x05
)

func getPath(params Parameters) (string, bool) {
	num, err := params.ReadAsString(PATH)
	if err != nil {
		return "", false
	}
	return num, true
}

func getMaxSubscribeID(params Parameters) (SubscribeID, bool) {
	num, err := params.ReadAsUint(MAX_SUBSCRIBE_ID)
	if err != nil {
		return 0, false
	}

	return SubscribeID(num), true
}

func getMaxCacheDuration(params Parameters) (time.Duration, bool) {
	num, err := params.ReadAsUint(MAX_CACHE_DURATION)
	if err != nil {
		return 0, false
	}

	return time.Duration(num), true
}

func getAuthorizationInfo(params Parameters) (string, bool) {
	str, err := params.ReadAsString(AUTHORIZATION_INFO)
	if err != nil {
		return "", false
	}

	return str, true
}

func getDeliveryTimeout(params Parameters) (time.Duration, bool) {
	num, err := params.ReadAsUint(DELIVERY_TIMEOUT)
	if err != nil {
		return 0, false
	}

	return time.Duration(num), true
}
