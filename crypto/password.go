package crypto

import (
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(p string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(p), 15)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func TestMatch(p string, h string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(h), []byte(p))
	if err != nil {
		return false
	}
	return true
}
