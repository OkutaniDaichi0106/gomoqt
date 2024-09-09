package moqtransport

import (
	"errors"
	"reflect"

	"github.com/quic-go/quic-go/quicvarint"
)

// Parameter
type ParameterKey uint64

const (
	ROLE               ParameterKey = 0x00
	PATH               ParameterKey = 0x01
	AUTHORIZATION_INFO ParameterKey = 0x02
	DELIVERY_TIMEOUT   ParameterKey = 0x03
	MAX_CACHE_DURATION ParameterKey = 0x04
	MAX_SUBSCRIBE_ID   ParameterKey = 0x05
	TRACK_NAME         ParameterKey = 0xf3 // Original
)

type WireType byte

const (
	/*
	 * varint indicates the following data is int, uint or bool
	 */
	varint WireType = 0

	/*
	 * length_delimited indicates the following data is byte array or string
	 */
	length_delimited WireType = 2
)

// Roles

type Role byte

const (
	PUB     Role = 0x00
	SUB     Role = 0x01
	PUB_SUB Role = 0x02
)

/*
 *
 *
 */

func (params Parameters) Role() (Role, error) {
	num, err := params.AsUint(ROLE)
	if errors.Is(err, ErrParameterNotFound) {
		return 0, ErrRoleNotFound
	}
	switch Role(num) {
	case PUB, SUB, PUB_SUB:
		return Role(num), nil
	default:
		return 0, ErrInvalidRole
	}
}

func (params Parameters) Path() (string, error) {
	num, err := params.AsString(PATH)
	if errors.Is(err, ErrParameterNotFound) {
		return "", ErrPathNotFound
	}
	return num, nil
}

func (params Parameters) MaxSubscribeID() (subscribeID, error) {
	num, err := params.AsUint(MAX_SUBSCRIBE_ID)
	if errors.Is(err, ErrParameterNotFound) {
		return 0, ErrMaxSubscribeIDNotFound
	}

	return subscribeID(num), nil
}

var ErrParameterNotFound = errors.New("parameter not found")

var ErrRoleNotFound = errors.New("role not found")
var ErrPathNotFound = errors.New("path not found")
var ErrMaxSubscribeIDNotFound = errors.New("max subscribe id not found")

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

var ErrNotBoolParameter = errors.New("it is assumed to not be a bool")
var ErrNotIntParameter = errors.New("it is assumed to not be a integer")
var ErrNotUintParameter = errors.New("it is assumed to not be a unsigned integer")
var ErrNotStringParameter = errors.New("it is assumed to not be a unsigned integer")
var ErrNotByteArrayParameter = errors.New("it is assumed to not be a unsigned integer")

/*
 * Parameters
 * Keys of the maps should not be duplicated
 */
type Parameters map[ParameterKey]any

func (params Parameters) AddParameter(key ParameterKey, value any) error {
	v, ok := params[key]

	// Check if the type of the existing value is the type of given value
	if ok && reflect.TypeOf(value) != reflect.TypeOf(v) {
		errors.New("you attempted to change an existing value to a different type ")
	}

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
	default:
		panic("invalid type")
	}

	return nil
}

func (params Parameters) serialize() []byte {
	data := make([]byte, 1<<4)

	for key, value := range params {
		switch v := value.(type) {
		case uint64:
			data = quicvarint.Append(data, uint64(key))
			data = quicvarint.Append(data, uint64(varint))
			data = quicvarint.Append(data, v)
		case []byte:
			data = quicvarint.Append(data, uint64(key))
			data = quicvarint.Append(data, uint64(length_delimited))
			data = quicvarint.Append(data, uint64(len(v)))
			data = append(data, v...)
		default:
			panic("invalid type")
		}
	}

	return data
}

func (params Parameters) parse(r quicvarint.Reader) error {
	var num uint64
	var err error
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
		// Get the uint64 data
		num, err = quicvarint.Read(r)
		if err != nil {
			return err
		}

		// Register the data
		params[key] = num
	case length_delimited:
		// Get length of the byte array data
		num, err = quicvarint.Read(r)
		if err != nil {
			return err
		}

		// Get byte array data
		buf := make([]byte, num)
		n, err := r.Read(buf)
		if err != nil {
			return err
		}

		// Register the data
		params[key] = buf[:n]
	default:
		return errors.New("invalid wire type")
	}

	return nil
}
