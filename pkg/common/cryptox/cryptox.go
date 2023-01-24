package cryptox

import (
	"crypto/aes"
	"crypto/sha256"
	"encoding/hex"
	log "github.com/sirupsen/logrus"
)

func EncryptString(key []byte, text string) string {
	c, err := aes.NewCipher(key)
	if err != nil {
		log.Fatalf("encryption key/salt is invalid: %s", err)
	}
	encrypted := make([]byte, len(text))
	c.Encrypt(encrypted, []byte(text))
	return hex.EncodeToString(encrypted)
}

func DecryptString(key []byte, encrypted string) string {
	c, err := aes.NewCipher(key)
	if err != nil {
		log.Fatalf("decryption key/salt is invalid: %s", err)
	}
	decoded, _ := hex.DecodeString(encrypted)
	dest := make([]byte, len(decoded))
	c.Decrypt(dest, decoded)
	s := string(dest[:])
	return s
}

func HashString(text string) string {
	h := sha256.New()
	h.Write([]byte(text))
	bs := h.Sum(nil)
	return string(bs)
}
