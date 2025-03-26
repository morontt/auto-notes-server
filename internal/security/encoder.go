package security

import (
	"golang.org/x/crypto/bcrypt"
)

func PasswordVerify(password, hash string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return false
	}

	return true
}
