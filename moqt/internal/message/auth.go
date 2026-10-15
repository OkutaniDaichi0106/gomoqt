package message

import (
	"errors"
	"io"
	"unicode/utf8"
)

// MaxAuthExpires is the largest AUTH_OK Expires, in milliseconds: 2^53 - 1.
const MaxAuthExpires = 1<<53 - 1

// MaxAuthReasonLength is the longest AUTH_ERROR Reason Phrase, in bytes.
const MaxAuthReasonLength = 8192

// ErrInvalidAuthMessage is returned when an Auth Stream message decodes but
// breaks a rule of its fields, which is a protocol violation.
var ErrInvalidAuthMessage = errors.New("invalid auth message")

// AuthReplyType is the Type that leads each message the acceptor sends on an
// Auth Stream.
type AuthReplyType uint64

const (
	AuthReplyOK    AuthReplyType = 0x0
	AuthReplyError AuthReplyType = 0x1
)

// ReadAuthReplyType reads the Type of the next acceptor message on an Auth
// Stream. The message itself follows, read with AuthOkMessage.Decode or
// AuthErrorMessage.Decode.
func ReadAuthReplyType(r io.Reader) (AuthReplyType, error) {
	t, err := ReadMessageLength(r)
	if err != nil {
		return 0, err
	}
	return AuthReplyType(t), nil
}

/*
 *	AUTH Message {
 *	  Message Length (i)
 *	  Token (b)
 *	}
 */
type AuthMessage struct {
	Token []byte
}

func (am AuthMessage) Len() int {
	return BytesLen(am.Token)
}

func (am AuthMessage) Encode(w io.Writer) error {
	msgLen := am.Len()
	b := make([]byte, 0, msgLen+VarintLen(uint64(msgLen)))

	b, _ = WriteMessageLength(b, uint64(msgLen))
	b, _ = WriteBytes(b, am.Token)

	_, err := w.Write(b)
	return err
}

func (am *AuthMessage) Decode(src io.Reader) error {
	b, err := readBody(src)
	if err != nil {
		return err
	}

	token, n, err := ReadBytes(b)
	if err != nil {
		return err
	}
	am.Token = append([]byte(nil), token...)
	b = b[n:]

	if len(b) != 0 {
		return ErrMessageTooShort
	}

	return nil
}

/*
 *	AUTH_OK Message {
 *	  Type (i) = 0x0
 *	  Message Length (i)
 *	  Publish Count (i)
 *	  Publish Pattern (s) ...
 *	  Subscribe Count (i)
 *	  Subscribe Pattern (s) ...
 *	  Expires (i)
 *	}
 */
type AuthOkMessage struct {
	PublishPatterns   []string
	SubscribePatterns []string
	// Expires is the number of milliseconds until the grant lapses, or 0 for
	// never.
	Expires uint64
}

func (aom AuthOkMessage) Len() int {
	return StringArrayLen(aom.PublishPatterns) + StringArrayLen(aom.SubscribePatterns) + VarintLen(aom.Expires)
}

// Encode writes the message with its leading Type.
func (aom AuthOkMessage) Encode(w io.Writer) error {
	if aom.Expires > MaxAuthExpires {
		return ErrInvalidAuthMessage
	}

	msgLen := aom.Len()
	b := make([]byte, 0, 1+msgLen+VarintLen(uint64(msgLen)))

	b, _ = WriteVarint(b, uint64(AuthReplyOK))
	b, _ = WriteMessageLength(b, uint64(msgLen))
	b, _ = WriteStringArray(b, aom.PublishPatterns)
	b, _ = WriteStringArray(b, aom.SubscribePatterns)
	b, _ = WriteVarint(b, aom.Expires)

	_, err := w.Write(b)
	return err
}

// Decode reads the message after its Type (see ReadAuthReplyType).
func (aom *AuthOkMessage) Decode(src io.Reader) error {
	b, err := readBody(src)
	if err != nil {
		return err
	}

	publish, n, err := ReadStringArray(b)
	if err != nil {
		return err
	}
	aom.PublishPatterns = publish
	b = b[n:]

	subscribe, n, err := ReadStringArray(b)
	if err != nil {
		return err
	}
	aom.SubscribePatterns = subscribe
	b = b[n:]

	expires, n, err := ReadVarint(b)
	if err != nil {
		return err
	}
	aom.Expires = expires
	b = b[n:]

	if len(b) != 0 {
		return ErrMessageTooShort
	}

	if aom.Expires > MaxAuthExpires {
		return ErrInvalidAuthMessage
	}

	return nil
}

/*
 *	AUTH_ERROR Message {
 *	  Type (i) = 0x1
 *	  Message Length (i)
 *	  Error Code (i)
 *	  Reason Phrase (s)
 *	}
 */
type AuthErrorMessage struct {
	// Code is a session error code, such as UNAUTHORIZED.
	Code   uint64
	Reason string
}

func (aem AuthErrorMessage) Len() int {
	return VarintLen(aem.Code) + StringLen(aem.Reason)
}

// Encode writes the message with its leading Type.
func (aem AuthErrorMessage) Encode(w io.Writer) error {
	if !aem.valid() {
		return ErrInvalidAuthMessage
	}

	msgLen := aem.Len()
	b := make([]byte, 0, 1+msgLen+VarintLen(uint64(msgLen)))

	b, _ = WriteVarint(b, uint64(AuthReplyError))
	b, _ = WriteMessageLength(b, uint64(msgLen))
	b, _ = WriteVarint(b, aem.Code)
	b, _ = WriteString(b, aem.Reason)

	_, err := w.Write(b)
	return err
}

// Decode reads the message after its Type (see ReadAuthReplyType).
func (aem *AuthErrorMessage) Decode(src io.Reader) error {
	b, err := readBody(src)
	if err != nil {
		return err
	}

	code, n, err := ReadVarint(b)
	if err != nil {
		return err
	}
	aem.Code = code
	b = b[n:]

	reason, n, err := ReadString(b)
	if err != nil {
		return err
	}
	aem.Reason = reason
	b = b[n:]

	if len(b) != 0 {
		return ErrMessageTooShort
	}

	if !aem.valid() {
		return ErrInvalidAuthMessage
	}

	return nil
}

func (aem AuthErrorMessage) valid() bool {
	return aem.Code <= 1<<32-1 && len(aem.Reason) <= MaxAuthReasonLength && utf8.ValidString(aem.Reason)
}

// readBody reads a length-prefixed message body.
func readBody(src io.Reader) ([]byte, error) {
	num, err := ReadMessageLength(src)
	if err != nil {
		return nil, err
	}

	if num > MaxMessageSize {
		return nil, ErrMessageTooLarge
	}

	b := make([]byte, num)
	if _, err := io.ReadFull(src, b); err != nil {
		return nil, err
	}
	return b, nil
}
