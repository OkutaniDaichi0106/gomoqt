package message

import (
	"io"

	"github.com/quic-go/quic-go/quicvarint"
)

/*
* Parameters
 */
type Parameters map[uint64][]byte

func appendParameters(b []byte, params Parameters) []byte {
	// Append the number of the parameters
	b = quicvarint.Append(b, uint64(len(params)))

	// Append the parameters
	for key, value := range params {
		// Append the Paramter Type
		b = quicvarint.Append(b, key)
		// Append the Paramter Length
		b = quicvarint.Append(b, uint64(len(value)))
		// Append the Paramter Value
		b = append(b, value...)
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

	for i := uint64(0); i < len; i++ {
		// Get a Parameter Type
		num, err := quicvarint.Read(r)
		if err != nil {
			return Parameters{}, err
		}
		key := num

		// Get a Parameter Length
		num, err = quicvarint.Read(r)
		if err != nil {
			return Parameters{}, err
		}

		// Get a Parameter Value
		buf := make([]byte, num)
		_, err = r.Read(buf)
		if err != nil {
			return Parameters{}, err
		}

		// Add the key and the value to the paramter map
		params[key] = buf
	}

	return params, nil
}
