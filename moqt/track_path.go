package moqt

import "strings"

type TrackPath string

// func NewTrackPath(path string) TrackPath {
// 	return TrackPath(path)
// }

func (tp TrackPath) String() string {
	return string(tp)
}

func (tp TrackPath) HasPrefix(prefix string) bool {
	if len(tp) < len(prefix) {
		return false
	}

	return strings.HasPrefix(string(tp), prefix+"/")
}

func (tp TrackPath) GetSuffix(prefix string) string {
	if !tp.HasPrefix(prefix) {
		return ""
	}

	return strings.TrimPrefix(string(tp), prefix+"/")
}

func (tp TrackPath) HasSuffix(suffix string) bool {
	if len(tp) < len(suffix) {
		return false
	}

	return strings.HasSuffix(string(tp), "/"+suffix)
}

func (tp TrackPath) Equal(target TrackPath) bool {
	return tp == target
}

func (tp TrackPath) Match(pattern string) bool {
	return matchGlob(pattern, string(tp))
}

func (tp TrackPath) Parts() []string {
	return strings.Split(string(tp), "/")
}
