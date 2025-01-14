package moqt

import (
	"bytes"
	"errors"
	"log/slog"

	"github.com/OkutaniDaichi0106/gomoqt/internal/message"
	"github.com/quic-go/quic-go/quicvarint"
)

type ParameterType uint64

type Parameters struct {
	paramMap message.Parameters
}

func NewParameters() Parameters {
	return Parameters{
		paramMap: make(message.Parameters),
	}
}

func (p Parameters) SetByteArray(key ParameterType, value []byte) {
	p.paramMap[uint64(key)] = value
}

func (p Parameters) SetString(key ParameterType, value string) {
	p.paramMap[uint64(key)] = []byte(value)
}

func (p Parameters) SetInt(key ParameterType, value int64) {
	p.paramMap[uint64(key)] = quicvarint.Append(make([]byte, 0), uint64(value))
}

func (p Parameters) SetUint(key ParameterType, value uint64) {
	p.paramMap[uint64(key)] = quicvarint.Append(make([]byte, 0), value)
}

func (p Parameters) SetBool(key ParameterType, value bool) {
	if value {
		p.paramMap[uint64(key)] = quicvarint.Append(make([]byte, 0), 1)
	} else {
		p.paramMap[uint64(key)] = quicvarint.Append(make([]byte, 0), 0)
	}
}

func (p Parameters) Remove(key ParameterType) {
	delete(p.paramMap, uint64(key))
}

func (p Parameters) GetByteArray(key ParameterType) ([]byte, error) {
	value, ok := p.paramMap[uint64(key)]
	if !ok {
		return nil, ErrParameterNotFound
	}

	return value, nil
}

func (p Parameters) GetString(key ParameterType) (string, error) {
	value, err := p.GetByteArray(key)
	if err != nil {
		slog.Error("failed to read a parameter as byte array")
		return "", err
	}

	return string(value), nil
}

func (p Parameters) GetInt(key ParameterType) (int64, error) {
	num, err := p.GetUint(key)
	if err != nil {
		slog.Error("failed to read a parameter as uint", slog.String("error", err.Error()))
		return 0, err
	}

	return int64(num), nil
}

func (p Parameters) GetUint(key ParameterType) (uint64, error) {
	value, ok := p.paramMap[uint64(key)]
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

func (p Parameters) GetBool(key ParameterType) (bool, error) {
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
	path               ParameterType = 0x01
	max_subscribe_id   ParameterType = 0x02
	authorization_info ParameterType = 0x03
	delivery_timeout   ParameterType = 0x04
	max_cache_duration ParameterType = 0x05
)

// // Path parameter
// func (p Parameters) SetPath(value string) {
// 	p.SetString(path, value)
// }

// func (p Parameters) GetPath() (string, bool) {
// 	num, err := p.GetString(path)
// 	if err != nil {
// 		return "", false
// 	}
// 	return num, true
// }

// // MaxSubscribeID parameter
// func (p Parameters) SetMaxSubscribeID(value SubscribeID) {
// 	p.SetUint(max_subscribe_id, uint64(value))
// }

// func (p Parameters) GetMaxSubscribeID() (SubscribeID, bool) {
// 	num, err := p.GetUint(max_subscribe_id)
// 	if err != nil {
// 		return 0, false
// 	}

// 	return SubscribeID(num), true
// }

// // MaxCacheDuration parameter
// func (p Parameters) SetMaxCacheDuration(value time.Duration) {
// 	p.SetInt(max_cache_duration, int64(value))
// }

// func (p Parameters) GetMaxCacheDuration() (time.Duration, bool) {
// 	num, err := p.GetUint(max_cache_duration)
// 	if err != nil {
// 		return 0, false
// 	}

// 	return time.Duration(num), true
// }

// // AuthorizationInfo parameter
// func (p Parameters) SetAuthorizationInfo(value string) {
// 	p.SetString(authorization_info, value)
// }

// func (p Parameters) GetAuthorizationInfo() (string, bool) {
// 	str, err := p.GetString(authorization_info)
// 	if err != nil {
// 		return "", false
// 	}

// 	return str, true
// }

// // DeliveryTimeout parameter
// func (p Parameters) SetDeliveryTimeout(value time.Duration) {
// 	p.SetInt(delivery_timeout, int64(value))
// }

// func (p Parameters) GetDeliveryTimeout() (time.Duration, bool) {
// 	num, err := p.GetUint(delivery_timeout)
// 	if err != nil {
// 		return 0, false
// 	}

// 	return time.Duration(num), true
// }
