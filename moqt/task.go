package moqt

import "strings"

func TrackPartsString(trackPath []string) string {
	var sb strings.Builder
	sb.WriteString("[")
	for i, path := range trackPath {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(path)
	}
	sb.WriteString("]")
	return sb.String()
}

type scheduler struct {
}

func (s *scheduler) Add() {}
