package cryptoutil

import (
	"crypto/rand"

	"github.com/pkg/errors"
)

const (
	// First-character set: leading char of a DNS label must be a letter for
	// maximum resolver/browser compatibility (RFC 1035; RFC 1123 relaxed this
	// but some legacy stacks still care).
	subdomainLeadingChars = "abcdefghijklmnopqrstuvwxyz"
	// Subsequent-character set: lowercase alphanumeric.
	subdomainChars = "abcdefghijklmnopqrstuvwxyz0123456789"
)

// RandomSubdomain returns a DNS-safe label of the given length composed of
// cryptographically random characters. The first character is always a letter;
// the remainder are lowercase alphanumeric.
//
// Rejection sampling is used to avoid modulo bias.
func RandomSubdomain(length int) (string, error) {
	if length < 1 {
		return "", errors.New("length must be >= 1")
	}
	out := make([]byte, length)
	first, err := randomByteFrom(subdomainLeadingChars)
	if err != nil {
		return "", err
	}
	out[0] = first
	for i := 1; i < length; i++ {
		b, err := randomByteFrom(subdomainChars)
		if err != nil {
			return "", err
		}
		out[i] = b
	}
	return string(out), nil
}

// randomByteFrom returns a single uniformly-distributed byte chosen from the
// given alphabet, using rejection sampling to avoid modulo bias.
func randomByteFrom(alphabet string) (byte, error) {
	n := byte(len(alphabet))
	// Largest multiple of n that fits in a byte; values above this are
	// rejected to keep the distribution uniform.
	max := byte(256/int(n)) * n
	buf := make([]byte, 1)
	for {
		if _, err := rand.Read(buf); err != nil {
			return 0, errors.Wrap(err, "read random")
		}
		if buf[0] < max {
			return alphabet[buf[0]%n], nil
		}
	}
}
