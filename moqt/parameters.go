package moqt

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"strings"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/internal/message"
	"github.com/quic-go/quic-go/quicvarint"
)

// ParameterType represents the type identifier for MOQ protocol parameters.
type ParameterType uint64

// NewParameters creates a new empty Parameters instance.
func NewParameters() *Parameters {
	return &Parameters{
		paramMap: make(message.Parameters),
	}
}

// Parameters holds key-value pairs for MOQ protocol negotiation.
// Parameters are used during session setup and other protocol operations
// to exchange configuration options between client and server.
type Parameters struct {
	paramMap message.Parameters
}

// Clone creates a deep copy of the Parameters.
func (p *Parameters) Clone() *Parameters {
	return &Parameters{
		paramMap: maps.Clone(p.paramMap),
	}
}

// String returns a string representation of the parameters for debugging.
func (p Parameters) String() string {
	var sb strings.Builder
	sb.WriteString("{")
	for key, value := range p.paramMap {
		sb.WriteString(" ")
		sb.WriteString(fmt.Sprintf("%d", key))
		sb.WriteString(": ")
		sb.WriteString(fmt.Sprintf("%v", value))
		sb.WriteString(",")
	}
	sb.WriteString(" }")
	return sb.String()
}

// SetByteArray sets a parameter with a byte array value.
func (p *Parameters) SetByteArray(key ParameterType, value []byte) {
	if p.paramMap == nil {
		p.paramMap = make(message.Parameters)
	}
	p.paramMap[uint64(key)] = value
}

// SetString sets a parameter with a string value.
func (p *Parameters) SetString(key ParameterType, value string) {
	if p.paramMap == nil {
		p.paramMap = make(message.Parameters)
	}
	p.paramMap[uint64(key)] = []byte(value)
}

// SetUint sets a parameter with an unsigned integer value encoded as a varint.
func (p *Parameters) SetUint(key ParameterType, value uint64) {
	if p.paramMap == nil {
		p.paramMap = make(message.Parameters)
	}
	p.paramMap[uint64(key)] = quicvarint.Append(make([]byte, 0), value)
}

// SetBool sets a parameter with a boolean value (1 for true, 0 for false).
func (p *Parameters) SetBool(key ParameterType, value bool) {
	if p.paramMap == nil {
		p.paramMap = make(message.Parameters)
	}
	if value {
		p.paramMap[uint64(key)] = quicvarint.Append(make([]byte, 0), 1)
	} else {
		p.paramMap[uint64(key)] = quicvarint.Append(make([]byte, 0), 0)
	}
}

// Remove removes a parameter by key.
func (p *Parameters) Remove(key ParameterType) {
	if p.paramMap == nil {
		return
	}
	delete(p.paramMap, uint64(key))
}

// GetByteArray retrieves a parameter value as a byte array.
// Returns ErrParameterNotFound if the parameter does not exist.
func (p Parameters) GetByteArray(key ParameterType) ([]byte, error) {
	if p.paramMap == nil {
		return nil, ErrParameterNotFound
	}
	value, ok := p.paramMap[uint64(key)]
	if !ok {
		return nil, ErrParameterNotFound
	}

	return value, nil
}

// GetString retrieves a parameter value as a string.
// Returns ErrParameterNotFound if the parameter does not exist.
func (p Parameters) GetString(key ParameterType) (string, error) {
	if p.paramMap == nil {
		return "", ErrParameterNotFound
	}

	value, err := p.GetByteArray(key)
	if err != nil {
		return "", err
	}

	return string(value), nil
}

// GetUint retrieves a parameter value as an unsigned integer decoded from a varint.
// Returns ErrParameterNotFound if the parameter does not exist.
func (p Parameters) GetUint(key ParameterType) (uint64, error) {
	if p.paramMap == nil {
		return 0, ErrParameterNotFound
	}

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

// GetBool retrieves a parameter value as a boolean (1=true, 0=false).
// Returns ErrParameterNotFound if the parameter does not exist.
func (p Parameters) GetBool(key ParameterType) (bool, error) {
	if p.paramMap == nil {
		return false, ErrParameterNotFound
	}

	num, err := p.GetUint(key)
	if err != nil {
		slog.Error("failed to read a parameter as uint", "error", err)
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
	param_type_path               ParameterType = 0x01
	param_type_authorization_info ParameterType = 0x02
	// param_type_delivery_timeout   ParameterType = 0x03
	// param_type_new_session_uri ParameterType = 0x04

	// max_subscribe_id   ParameterType = 0x02
)
