package cryptox

import (
	"crypto/aes"
	"encoding/hex"
	log "github.com/sirupsen/logrus"
)

func Encrypt(key []byte, plaintext string) (string, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		log.Fatalf("encryption key/salt is invalid: %s", err)
	}
	encrypted := make([]byte, len(plaintext))
	c.Encrypt(encrypted, []byte(plaintext))
	return hex.EncodeToString(encrypted), nil
}

func Decrypt(key []byte, ct string) (string, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		log.Fatalf("encryption key/salt is invalid: %s", err)
	}
	decoded, _ := hex.DecodeString(ct)
	dest := make([]byte, len(decoded))
	c.Decrypt(dest, decoded)
	s := string(dest[:])
	return s, nil
}
