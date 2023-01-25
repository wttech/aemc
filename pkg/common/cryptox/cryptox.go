package cryptox

import (
	"crypto/aes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
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
	hash := sha256.New()
	_, _ = io.WriteString(hash, text)
	return fmt.Sprintf("%x", hash.Sum(nil))
}
