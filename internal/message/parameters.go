package message

import (
	"errors"
	"io"

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

func readParameters(r reader) (Parameters, error) {
	// Get the number of the parameters
	len, err := quicvarint.Read(r)
	if err != nil {
		if err == io.EOF {
			return make(Parameters), nil
		}
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
