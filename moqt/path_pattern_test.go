package moqt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePathPattern(t *testing.T) {
	tests := map[string]struct {
		input   string
		wantErr bool
	}{
		"root":                    {input: "/"},
		"path":                    {input: "/room/1/alice"},
		"prefix":                  {input: "/room/1/"},
		"empty":                   {input: "", wantErr: true},
		"no leading /":            {input: "room/1", wantErr: true},
		"repeated /":              {input: "/room//1", wantErr: true},
		"repeated root /":         {input: "//", wantErr: true},
		"wildcard":                {input: "/room/{id}/host"},
		"wildcard prefix":         {input: "/room/{id}/"},
		"rest wildcard":           {input: "/room/{rest...}"},
		"two wildcards":           {input: "/room/{id}/{name}"},
		"wildcard inside segment": {input: "/room/a{id}", wantErr: true},
		"unclosed wildcard":       {input: "/room/{id", wantErr: true},
		"stray closing brace":     {input: "/room/id}", wantErr: true},
		"empty wildcard name":     {input: "/room/{}", wantErr: true},
		"non-identifier name":     {input: "/room/{1d}", wantErr: true},
		"duplicate name":          {input: "/{id}/{id}", wantErr: true},
		"rest wildcard not last":  {input: "/{rest...}/host", wantErr: true},
		"rest wildcard then /":    {input: "/{rest...}/", wantErr: true},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			p, err := ParsePathPattern(tt.input)
			if tt.wantErr {
				assert.ErrorIs(t, err, ErrInvalidPathPattern)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.input, p.String())
		})
	}
}

func TestPathPattern_Match(t *testing.T) {
	tests := map[string]struct {
		pattern string
		path    BroadcastPath
		want    bool
	}{
		"path matches itself":                  {pattern: "/room/1/alice", path: "/room/1/alice", want: true},
		"path matches no other":                {pattern: "/room/1/alice", path: "/room/1/bob", want: false},
		"path matches nothing beneath":         {pattern: "/room/1", path: "/room/1/alice", want: false},
		"prefix matches beneath":               {pattern: "/room/1/", path: "/room/1/alice", want: true},
		"prefix matches deeper":                {pattern: "/room/1/", path: "/room/1/a/b", want: true},
		"prefix stops at a / boundary":         {pattern: "/room/1/", path: "/room/12/alice", want: false},
		"prefix does not match its stem":       {pattern: "/room/1/", path: "/room/1", want: false},
		"root matches every path":              {pattern: "/", path: "/anything/at/all", want: true},
		"wildcard matches one segment":         {pattern: "/room/{id}/host", path: "/room/1/host", want: true},
		"wildcard needs the literal after it":  {pattern: "/room/{id}/host", path: "/room/1/guest", want: false},
		"wildcard matches no empty segment":    {pattern: "/room/{id}", path: "/room/", want: false},
		"wildcard matches no two segments":     {pattern: "/room/{id}", path: "/room/1/host", want: false},
		"wildcard path matches nothing deeper": {pattern: "/room/{id}/host", path: "/room/1/host/a", want: false},
		"wildcard prefix matches beneath":      {pattern: "/room/{id}/", path: "/room/1/alice", want: true},
		"wildcard prefix matches deeper":       {pattern: "/room/{id}/", path: "/room/1/a/b", want: true},
		"wildcard prefix needs a segment":      {pattern: "/room/{id}/", path: "/room/1", want: false},
		"wildcard needs its literal before":    {pattern: "/room/{id}/", path: "/lobby/1/alice", want: false},
		"rest wildcard matches one segment":    {pattern: "/room/{rest...}", path: "/room/1", want: true},
		"rest wildcard matches deeper":         {pattern: "/room/{rest...}", path: "/room/1/a/b", want: true},
		"rest wildcard needs a segment":        {pattern: "/room/{rest...}", path: "/room/", want: false},
		"rest wildcard needs its stem":         {pattern: "/room/{rest...}", path: "/room", want: false},
		"wildcard matches no invalid path":     {pattern: "/{id}", path: "room", want: false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			p := MustParsePathPattern(tt.pattern)

			assert.Equal(t, tt.want, p.Match(tt.path))
		})
	}
}

func TestPathPattern_Match_ZeroValue(t *testing.T) {
	var p PathPattern

	assert.False(t, p.Match("/"))
	assert.False(t, p.Match("/room"))
}
