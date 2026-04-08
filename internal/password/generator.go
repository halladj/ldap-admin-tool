package password

import (
	"crypto/rand"
	"math/big"
)

const (
	lowerChars   = "abcdefghijklmnopqrstuvwxyz"
	upperChars   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digitChars   = "0123456789"
	specialChars = "!@#$%&*"
	allChars     = lowerChars + upperChars + digitChars + specialChars
)

func Generate(length int) (string, error) {
	if length < 8 {
		length = 12
	}

	for {
		password := make([]byte, length)
		for i := range password {
			idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(allChars))))
			if err != nil {
				return "", err
			}
			password[i] = allChars[idx.Int64()]
		}

		if isStrong(string(password)) {
			return string(password), nil
		}
	}
}

func isStrong(p string) bool {
	var hasLower, hasUpper, hasDigit, hasSpecial bool
	for _, c := range p {
		switch {
		case c >= 'a' && c <= 'z':
			hasLower = true
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= '0' && c <= '9':
			hasDigit = true
		default:
			hasSpecial = true
		}
	}
	return hasLower && hasUpper && hasDigit && hasSpecial
}
