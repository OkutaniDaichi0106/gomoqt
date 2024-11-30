package message

import (
	"errors"

	"github.com/quic-go/quic-go/quicvarint"
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
type Parameters map[uint64]any

func (params Parameters) AsBool(key uint64) (bool, bool) {
	value, ok := params[key]
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

func (params Parameters) AsInt(key uint64) (int64, bool) {
	value, ok := params[key]
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

func (params Parameters) AsUint(key uint64) (uint64, bool) {
	value, ok := params[key]
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

func (params Parameters) AsString(key uint64) (string, bool) {
	value, ok := params.AsByteArray(key)
	if !ok {
		return "", false
	}
	return string(value), true
}

func (params Parameters) AsByteArray(key uint64) ([]byte, bool) {
	value, ok := params[key]
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

func (params Parameters) Remove(key uint64) {
	delete(params, key)
}

func (params Parameters) Add(key uint64, value any) {
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
}

func appendParameters(b []byte, params Parameters) []byte {
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

func readParameters(r Reader) (Parameters, error) {
	// Get the number of the parameters
	len, err := quicvarint.Read(r)
	if err != nil {
		return Parameters{}, err
	}

	params := make(Parameters, len)

	var num uint64
	for i := uint64(0); i < len; i++ {
		num, err = quicvarint.Read(r)
		if err != nil {
			return Parameters{}, err
		}
		key := uint64(num)

		num, err = quicvarint.Read(r)
		if err != nil {
			return Parameters{}, err
		}
		wireType := WireType(num)

		switch wireType {
		case varint:
			// Get the uint64 b
			num, err = quicvarint.Read(r)
			if err != nil {
				return Parameters{}, err
			}

			// Register the b
			params[key] = num
		case length_delimited:
			// Get length of the byte array b
			num, err = quicvarint.Read(r)
			if err != nil {
				return Parameters{}, err
			}

			// Get byte array b
			buf := make([]byte, num)
			n, err := r.Read(buf)
			if err != nil {
				return Parameters{}, err
			}

			// Register the b
			params[key] = buf[:n]
		default:
			return Parameters{}, errors.New("invalid wire type")
		}
	}

	return params, nil
}
