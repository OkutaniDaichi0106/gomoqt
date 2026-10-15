package moqt

import (
	"strings"
	"testing"
	"testing/synctest"
	"time"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthWriter_Grant_InvalidPattern(t *testing.T) {
	scope := newAuthScope(func() {})
	w := &AuthWriter{scope: scope}

	err := w.Grant(Grant{Subscribe: []string{"/room/1/", "room"}})

	assert.ErrorIs(t, err, ErrInvalidPathPattern)
	assert.False(t, scope.canSubscribe("/room/1"), "the invalid grant is not applied")
}

func TestAuthWriter_Grant_ReusedSlice(t *testing.T) {
	changes := 0
	scope := newAuthScope(func() { changes++ })
	w := &AuthWriter{scope: scope}
	g := Grant{Subscribe: []string{"/room/1/"}}
	require.NoError(t, w.Grant(g))

	g.Subscribe[0] = "/room/2/"
	require.NoError(t, w.Grant(g))

	assert.False(t, scope.canSubscribe("/room/1/host"))
	assert.True(t, scope.canSubscribe("/room/2/host"))
	assert.Equal(t, 2, changes, "editing the slice in place and granting again changes the scope")
}

func TestAuthWriter_Grant_MaxAge(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		scope := newAuthScope(func() {})
		w := &AuthWriter{scope: scope}
		require.NoError(t, w.Grant(Grant{Subscribe: []string{"/"}, MaxAge: time.Second}))

		time.Sleep(2 * time.Second)
		synctest.Wait()

		assert.False(t, scope.canSubscribe("/room/1"))
		assert.ErrorIs(t, w.Grant(Grant{}), ErrAuthEnded, "an expired grant ends its token")
	})
}

func TestAuthWriter_Grant_RegrantExtends(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		scope := newAuthScope(func() {})
		w := &AuthWriter{scope: scope}
		g := Grant{Subscribe: []string{"/"}, MaxAge: time.Second}
		require.NoError(t, w.Grant(g))

		time.Sleep(700 * time.Millisecond)
		require.NoError(t, w.Grant(g))
		time.Sleep(700 * time.Millisecond)
		synctest.Wait()

		assert.True(t, scope.canSubscribe("/room/1"), "the second grant restarts the expiry")
	})
}

func TestAuthScope_set_SamePatterns(t *testing.T) {
	changes := 0
	scope := newAuthScope(func() { changes++ })
	w := &AuthWriter{scope: scope}
	g, err := parseGrant(Grant{Subscribe: []string{"/room/1/"}})
	require.NoError(t, err)
	again, err := parseGrant(Grant{Subscribe: []string{"/room/1/"}, MaxAge: time.Minute})
	require.NoError(t, err)

	scope.set(w, g)
	scope.set(w, again)

	assert.Equal(t, 1, changes, "extending a grant does not change the scope")
}

func TestAuthReason(t *testing.T) {
	long := strings.Repeat("あ", maxAuthReasonLength)
	tests := map[string]struct {
		reason string
		want   string
	}{
		"valid":        {reason: "unknown ticket", want: "unknown ticket"},
		"invalid utf8": {reason: "bad\xffbyte", want: "bad�byte"},
		"too long":     {reason: long, want: long[:maxAuthReasonLength-maxAuthReasonLength%3]},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := authReason(tt.reason)

			assert.Equal(t, tt.want, got)
			assert.True(t, utf8.ValidString(got))
			assert.LessOrEqual(t, len(got), maxAuthReasonLength)
		})
	}
}
