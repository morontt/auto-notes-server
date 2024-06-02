package security

import (
	"crypto/sha512"
	"encoding/base64"
)

const iterations = 4600

func EncodePassword(password string, salt string) string {
	salted := []byte(password + "{" + salt + "}")
	h := sha512.New384()

	h.Write(salted)
	digest := h.Sum(nil)

	for i := 1; i < iterations; i++ {
		h.Reset()
		h.Write(digest)
		h.Write(salted)
		digest = h.Sum(nil)
	}

	return base64.StdEncoding.EncodeToString(digest)
}
