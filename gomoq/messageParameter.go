package gomoq

import (
	"errors"
	"io"

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
type Parameter struct {
	Key ParameterKey
	/*
	 * when WireType is VARINT,
	 *
	 * when WireType is VARINT,
	 */
	WireType

	// values
	/*
	 * value_int is only used when this WireType equals VARINT
	 * It not just means integer but also means flag with boolean
	 */
	value_int64 uint64

	/*
	 * value_string is only used when this WireType equals LENGTH_DELIMITED
	 */
	value_string string
}

func (p Parameter) append(b []byte) []byte {
	/*
	 * Integer Parameter {
	 *   Parameter Type (varint),
	 *   Number (varint),
	 * }
	 *
	 * String Parameter {
	 *   Parameter Type (varint),
	 *   Parameter Length (varint),
	 *   Parameter String ([]byte),
	 * }
	 */
	b = quicvarint.Append(b, uint64(p.Key))
	switch p.WireType {
	case varint:
		b = quicvarint.Append(b, uint64(varint))
		b = quicvarint.Append(b, p.value_int64)
	case length_delimited:
		b = quicvarint.Append(b, uint64(length_delimited))
		b = quicvarint.Append(b, uint64(len(p.value_string)))
		b = append(b, []byte(p.value_string)...)
	}

	return b
}

/*
 * Parameters
 * Keys of the maps should not be duplicated
 */
type Parameters []Parameter

/*
 * Add an int parameter to the Parameters
 * This function should not be used within the library
 * You should not add any role parameter in this function
 * because Role Parameter are automatically added by the Publiser or the Subscriber
 */
func (ps *Parameters) AddIntParameter(typeKey ParameterKey, num uint64) {
	// Avoid to add Role Parameter
	if typeKey == ROLE {
		panic("Role Parameter should not be added outside the internal system")
	}

	ps.addIntParameter(typeKey, num)
}

/*
 * This function should be used within the library
 */
func (ps *Parameters) addIntParameter(typeKey ParameterKey, num uint64) {
	*ps = append(*ps, Parameter{
		Key:         typeKey,
		WireType:    varint,
		value_int64: num,
	})
}

func (ps *Parameters) AddStringParameter(typeKey ParameterKey, str string) {
	*ps = append(*ps, Parameter{
		Key:          typeKey,
		WireType:     length_delimited,
		value_string: str,
	})
}

func (ps *Parameters) AddBoolParameter(typeKey ParameterKey, flag bool) {
	/*
	 * Value {
	 *   Flag (0 or 1),
	 * }
	 *
	 * false is stored as 0, true is stored as 1 in Parameter.Value
	 */
	if !flag {
		ps.AddIntParameter(typeKey, 0)
	} else if flag {
		ps.AddIntParameter(typeKey, 1)
	} else {
		panic("the flag is neither false nor true")
	}
}

func (ps Parameters) append(b []byte) []byte {
	/*
	 * Parameters {
	 *   Number of Parameters (varint),
	 *   Parameter (..),
	 *   ...
	 * }
	 */
	// Append the number of the paramters
	b = quicvarint.Append(b, uint64(len(ps)))

	// Append serialized parameters
	for _, param := range ps {
		b = param.append(b)
	}

	return b
}

func (params *Parameters) parse(r quicvarint.Reader) error {
	var err error
	var param Parameter

	//Initialize parameters field
	*params = Parameters{}
	var (
		typeKey  uint64
		wireType uint64
	)

	num, err := quicvarint.Read(r)
	if err != nil {
		return err
	}

	for i := uint64(0); i < num; i++ {
		// Get parameter key
		typeKey, err = quicvarint.Read(r)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		// Get wire type
		wireType, err = quicvarint.Read(r)
		if err != nil {
			return err
		}
		// Parse the parameters
		switch WireType(wireType) {
		case varint:
			numv, err := quicvarint.Read(r)
			if err != nil {
				return err
			}
			param = Parameter{
				Key:         ParameterKey(typeKey),
				WireType:    WireType(wireType),
				value_int64: numv,
			}
		case length_delimited:
			length, err := quicvarint.Read(r)
			if err != nil {
				return err
			}
			buf := make([]byte, length)
			_, err = r.Read(buf)
			if err != nil {
				return err
			}
			param = Parameter{
				Key:          ParameterKey(typeKey),
				WireType:     WireType(wireType),
				value_string: string(buf),
			}
		default:
			return errors.New("invalid wire type")
		}
		*params = append(*params, param)
	}
	return nil
}

func (ps Parameters) Contain(key ParameterKey) (bool, any) {
	var param Parameter
	for _, param = range ps {
		if param.Key == key {
			switch param.WireType {
			case varint:
				return true, param.value_int64
			case length_delimited:
				return true, param.value_string
			}
			// Anything else varint or length_delimited is unacceptable as Wire Type
			return false, nil
		}
	}
	return false, nil
}
