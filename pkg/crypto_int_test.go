//go:build int_test

package pkg_test

import (
	"github.com/wttech/aemc/pkg"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCryptoProtect(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	aem := pkg.DefaultAEM()
	instance := aem.InstanceManager().NewLocalAuthor()

	protected, err := instance.Crypto().Protect("hello")
	a.Nil(err, "cannot protect value")
	a.NotEmpty(protected, "protected value should not be empty")
	a.Contains(protected, "{", "protected value should be wrapped in curly braces")
	a.Contains(protected, "}", "protected value should be wrapped in curly braces")
}

func TestCryptoUnprotect(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	aem := pkg.DefaultAEM()
	instance := aem.InstanceManager().NewLocalAuthor()

	protected, err := instance.Crypto().Protect("hello")
	a.Nil(err, "cannot protect value")

	plain, err := instance.Crypto().Unprotect(protected)
	a.Nil(err, "cannot unprotect value")
	a.Equal("hello", plain, "unprotected value should match original")
}

func TestCryptoUnprotectSpecialChars(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	aem := pkg.DefaultAEM()
	instance := aem.InstanceManager().NewLocalAuthor()

	values := []string{
		`he said "hello"`,
		`<script>alert("xss")&foo</script>`,
		`it's a test`,
		`jdbc:mysql://host:3306/db?user=admin&password=p@ss"w0rd`,
	}

	for _, original := range values {
		protected, err := instance.Crypto().Protect(original)
		a.Nil(err, "cannot protect value: %s", original)

		plain, err := instance.Crypto().Unprotect(protected)
		a.Nil(err, "cannot unprotect value: %s", original)
		a.Equal(original, plain, "roundtrip failed for value: %s", original)
	}
}

func TestCryptoUnprotectInvalidCiphertext(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	aem := pkg.DefaultAEM()
	instance := aem.InstanceManager().NewLocalAuthor()

	_, err := instance.Crypto().Unprotect("{invalid_cipher_text}")
	a.NotNil(err, "unprotecting invalid ciphertext should return an error")
	a.Contains(err.Error(), "decryption failed", "error should indicate decryption failure")
}
