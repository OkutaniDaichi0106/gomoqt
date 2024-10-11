package moqtmessage

import (
	"errors"
	"time"

	"github.com/quic-go/quic-go/quicvarint"
)

type Role byte

const (
	PUB     Role = 0x00
	SUB     Role = 0x01
	PUB_SUB Role = 0x02
)

// Parameter
type ParameterKey uint64

// Setup Parameters
const (
	ROLE             ParameterKey = 0x00
	PATH             ParameterKey = 0x01
	MAX_SUBSCRIBE_ID ParameterKey = 0x02
)

const (
	AUTHORIZATION_INFO ParameterKey = 0x02
	DELIVERY_TIMEOUT   ParameterKey = 0x03
	MAX_CACHE_DURATION ParameterKey = 0x04
)

type WireType byte

const (
	/*
	 * varint indicates the following b is int, uint or bool
	 */
	varint WireType = 0

	/*
	 * length_delimited indicates the following b is byte array or string
	 */
	length_delimited WireType = 2
)

/*
 * Parameters
 * Keys of the maps should not be duplicated
 */
type Parameters map[ParameterKey]any

func (params Parameters) Role() (Role, bool) {
	num, err := params.AsUint(ROLE)
	if err != nil {
		return 0, false
	}
	return Role(num), true
}

func (params Parameters) Path() (string, bool) {
	num, err := params.AsString(PATH)
	if err != nil {
		return "", false
	}
	return num, true
}

func (params Parameters) MaxSubscribeID() (SubscribeID, bool) {
	num, err := params.AsUint(MAX_SUBSCRIBE_ID)
	if err != nil {
		return 0, false
	}

	return SubscribeID(num), true
}

func (params Parameters) MaxCacheDuration() (time.Duration, bool) {
	num, err := params.AsUint(MAX_CACHE_DURATION)
	if err != nil {
		return 0, false
	}

	return time.Duration(num), true
}

func (params Parameters) AuthorizationInfo() (string, bool) {
	str, err := params.AsString(AUTHORIZATION_INFO)
	if err != nil {
		return "", false
	}

	return str, true
}

func (params Parameters) DeliveryTimeout() (time.Duration, bool) {
	num, err := params.AsUint(DELIVERY_TIMEOUT)
	if err != nil {
		return 0, false
	}

	return time.Duration(num), true
}

var ErrParameterNotFound = errors.New("parameter not found")

var ErrRoleNotFound = errors.New("role not found")
var ErrPathNotFound = errors.New("path not found")
var ErrMaxSubscribeIDNotFound = errors.New("max subscribe id not found")
var ErrMaxCacheDurationNotFound = errors.New("max cache duration not found")
var ErrAuthorizationInfoNotFound = errors.New("authorization info not found")
var ErrDeliveryTimeoutNotFound = errors.New("delivery timeout not found")

func (params Parameters) AsBool(key ParameterKey) (bool, error) {
	value, ok := params[key]
	if !ok {
		return false, ErrParameterNotFound
	}
	switch v := value.(type) {
	case uint64:
		if v == 0 {
			return false, nil
		} else if v == 1 {
			return true, nil
		} else {
			return false, ErrNotBoolParameter
		}
	default:
		return false, ErrNotBoolParameter
	}
}

func (params Parameters) AsInt(key ParameterKey) (int64, error) {
	value, ok := params[key]
	if !ok {
		return 0, ErrParameterNotFound
	}
	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case uint:
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint64:
		return int64(v), nil
	default:
		return 0, ErrNotIntParameter
	}
}

func (params Parameters) AsUint(key ParameterKey) (uint64, error) {
	value, ok := params[key]
	if !ok {
		return 0, ErrParameterNotFound
	}
	switch v := value.(type) {
	case int:
		return uint64(v), nil
	case int8:
		return uint64(v), nil
	case int16:
		return uint64(v), nil
	case int32:
		return uint64(v), nil
	case int64:
		return uint64(v), nil
	case uint:
		return uint64(v), nil
	case uint8:
		return uint64(v), nil
	case uint16:
		return uint64(v), nil
	case uint32:
		return uint64(v), nil
	case uint64:
		return v, nil
	default:
		return 0, ErrNotUintParameter
	}
}

func (params Parameters) AsString(key ParameterKey) (string, error) {
	value, err := params.AsByteArray(key)
	if value == nil || err != nil {
		if errors.Is(err, ErrNotByteArrayParameter) {
			err = ErrNotStringParameter
		}
		return "", err
	}
	return string(value), nil
}

func (params Parameters) AsByteArray(key ParameterKey) ([]byte, error) {
	value, ok := params[key]
	if !ok {
		return nil, ErrParameterNotFound
	}
	switch v := value.(type) {
	case []byte:
		return v, nil
	default:
		return nil, ErrNotByteArrayParameter
	}
}

func (params Parameters) Remove(key ParameterKey) {
	delete(params, key)
}

var ErrNotBoolParameter = errors.New("it is assumed to not be a bool")
var ErrNotIntParameter = errors.New("it is assumed to not be a integer")
var ErrNotUintParameter = errors.New("it is assumed to not be a unsigned integer")
var ErrNotStringParameter = errors.New("it is assumed to not be a unsigned integer")
var ErrNotByteArrayParameter = errors.New("it is assumed to not be a unsigned integer")

func (params Parameters) AddParameter(key ParameterKey, value any) {
	switch v := value.(type) {
	case int64:
		params[key] = uint64(v)
	case int32:
		params[key] = uint64(v)
	case int16:
		params[key] = uint64(v)
	case int8:
		params[key] = uint64(v)
	case uint32:
		params[key] = uint64(v)
	case uint16:
		params[key] = uint64(v)
	case uint8:
		params[key] = uint64(v)
	case uint64:
		params[key] = v
	case bool:
		if v {
			params[key] = 1
		} else if !v {
			params[key] = 0
		}
	case string:
		params[key] = []byte(v)
	case []byte:
		params[key] = v
	case Role:
		params[key] = uint64(v)
	case SubscribeID:
		params[key] = uint64(v)
	default:
		panic("invalid type")
	}

}

func (params Parameters) append(b []byte) []byte {
	// Append the number of the parameters
	b = quicvarint.Append(b, uint64(len(params)))

	// Append the parameters
	for key, value := range params {
		switch v := value.(type) {
		case uint64:
			b = quicvarint.Append(b, uint64(key))
			b = quicvarint.Append(b, uint64(varint))
			b = quicvarint.Append(b, v)
		case []byte:
			b = quicvarint.Append(b, uint64(key))
			b = quicvarint.Append(b, uint64(length_delimited))
			b = quicvarint.Append(b, uint64(len(v)))
			b = append(b, v...)
		default:
			panic("invalid type")
		}
	}

	return b
}

func (params *Parameters) Deserialize(r quicvarint.Reader) error {
	var num uint64
	var err error

	// Get the number of the parameters
	num, err = quicvarint.Read(r)
	if err != nil {
		return err
	}

	if *params == nil {
		*params = make(Parameters, int(num))
	}

	numParam := num

	for i := uint64(0); i < numParam; i++ {
		num, err = quicvarint.Read(r)
		if err != nil {
			return err
		}
		key := ParameterKey(num)

		num, err = quicvarint.Read(r)
		if err != nil {
			return err
		}
		wireType := WireType(num)

		switch wireType {
		case varint:
			// Get the uint64 b
			num, err = quicvarint.Read(r)
			if err != nil {
				return err
			}

			// Register the b
			(*params)[key] = num
		case length_delimited:
			// Get length of the byte array b
			num, err = quicvarint.Read(r)
			if err != nil {
				return err
			}

			// Get byte array b
			buf := make([]byte, num)
			n, err := r.Read(buf)
			if err != nil {
				return err
			}

			// Register the b
			(*params)[key] = buf[:n]
		default:
			return errors.New("invalid wire type")
		}
	}

	return nil
}
