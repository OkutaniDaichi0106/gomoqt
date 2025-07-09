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

type ParameterType uint64

func NewParameters() *Parameters {
	return &Parameters{
		paramMap: make(message.Parameters),
	}
}

type Parameters struct {
	paramMap message.Parameters
}

func (p *Parameters) Clone() *Parameters {
	return &Parameters{
		paramMap: maps.Clone(p.paramMap),
	}
}

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

func (p *Parameters) SetByteArray(key ParameterType, value []byte) {
	if p.paramMap == nil {
		p.paramMap = make(message.Parameters)
	}
	p.paramMap[uint64(key)] = value
}

func (p *Parameters) SetString(key ParameterType, value string) {
	if p.paramMap == nil {
		p.paramMap = make(message.Parameters)
	}
	p.paramMap[uint64(key)] = []byte(value)
}

func (p *Parameters) SetUint(key ParameterType, value uint64) {
	if p.paramMap == nil {
		p.paramMap = make(message.Parameters)
	}
	p.paramMap[uint64(key)] = quicvarint.Append(make([]byte, 0), value)
}

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

func (p *Parameters) Remove(key ParameterType) {
	if p.paramMap == nil {
		return
	}
	delete(p.paramMap, uint64(key))
}

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

func (p Parameters) GetString(key ParameterType) (string, error) {
	if p.paramMap == nil {
		return "", ErrParameterNotFound
	}

	value, err := p.GetByteArray(key)
	if err != nil {
		slog.Error("failed to read a parameter as byte array")
		return "", err
	}

	return string(value), nil
}

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
	param_type_new_session_uri ParameterType = 0x04

	// max_subscribe_id   ParameterType = 0x02
)
