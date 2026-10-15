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
		"root":            {input: "/"},
		"path":            {input: "/room/1/alice"},
		"prefix":          {input: "/room/1/"},
		"empty":           {input: "", wantErr: true},
		"no leading /":    {input: "room/1", wantErr: true},
		"repeated /":      {input: "/room//1", wantErr: true},
		"repeated root /": {input: "//", wantErr: true},
		"reserved brace":  {input: "/room/{id}/", wantErr: true},
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
		"path matches itself":            {pattern: "/room/1/alice", path: "/room/1/alice", want: true},
		"path matches no other":          {pattern: "/room/1/alice", path: "/room/1/bob", want: false},
		"path matches nothing beneath":   {pattern: "/room/1", path: "/room/1/alice", want: false},
		"prefix matches beneath":         {pattern: "/room/1/", path: "/room/1/alice", want: true},
		"prefix matches deeper":          {pattern: "/room/1/", path: "/room/1/a/b", want: true},
		"prefix stops at a / boundary":   {pattern: "/room/1/", path: "/room/12/alice", want: false},
		"prefix does not match its stem": {pattern: "/room/1/", path: "/room/1", want: false},
		"root matches every path":        {pattern: "/", path: "/anything/at/all", want: true},
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
