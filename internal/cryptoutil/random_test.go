package cryptoutil

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRandomSubdomain(t *testing.T) {
	pattern := regexp.MustCompile(`^[a-z][a-z0-9]*$`)

	t.Run("rejects non-positive lengths", func(t *testing.T) {
		_, err := RandomSubdomain(0)
		require.Error(t, err)
		_, err = RandomSubdomain(-1)
		require.Error(t, err)
	})

	t.Run("respects the requested length", func(t *testing.T) {
		for _, n := range []int{1, 5, 12, 32, 63} {
			s, err := RandomSubdomain(n)
			require.NoError(t, err)
			assert.Len(t, s, n)
		}
	})

	t.Run("only emits DNS-safe chars with a leading letter", func(t *testing.T) {
		for i := 0; i < 200; i++ {
			s, err := RandomSubdomain(16)
			require.NoError(t, err)
			assert.True(t, pattern.MatchString(s), "subdomain %q must match %s", s, pattern)
		}
	})

	t.Run("produces unique values across many invocations", func(t *testing.T) {
		seen := make(map[string]struct{}, 5000)
		for i := 0; i < 5000; i++ {
			s, err := RandomSubdomain(12)
			require.NoError(t, err)
			_, dup := seen[s]
			require.False(t, dup, "unexpected duplicate %q at iteration %d", s, i)
			seen[s] = struct{}{}
		}
	})
}
