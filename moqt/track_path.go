package moqt

import "strings"

type TrackPath []string

func NewTrackPath(path []string) TrackPath {
	return TrackPath(path)
}

func (tp TrackPath) String() string {
	var sb strings.Builder

	sb.WriteString("[")
	for i, path := range tp {
		if i > 0 {
			sb.WriteString(", ")
		}

		sb.WriteString(path)
	}
	sb.WriteString("]")
	return sb.String()
}

func (tp TrackPath) HasPrefix(prefix []string) bool {
	if len(tp) < len(prefix) {
		return false
	}
	for i, v := range prefix {
		if tp[i] != v {
			return false
		}
	}
	return true
}

func (tp TrackPath) GetSuffix(prefix []string) []string {
	if !tp.HasPrefix(prefix) {
		return nil
	}
	return tp[len(prefix):]
}

func (tp TrackPath) HasSuffix(suffix []string) bool {
	if len(tp) < len(suffix) {
		return false
	}
	for i, v := range suffix {
		if tp[len(tp)-len(suffix)+i] != v {
			return false
		}
	}
	return true
}

func (tp TrackPath) Equal(other TrackPath) bool {
	if len(tp) != len(other) {
		return false
	}
	for i, v := range tp {
		if v != other[i] {
			return false
		}
	}
	return true
}
