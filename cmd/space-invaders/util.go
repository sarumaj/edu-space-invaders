package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"net/url"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

// associatedData is the associated data for the AES GCM cipher.
const associatedData = "aes256gcm"

// decodeB64AndDecryptWithAES decodes the base64-encoded message and decrypts it using the AES key.
func decodeB64AndDecryptWithAES(keyCipher cipher.AEAD, b64encoded string) (string, error) {
	// Trim the padding characters from the encrypted message
	encryptedMessage := strings.TrimSuffix(b64encoded, string(base64.StdPadding))

	// Decode the encrypted message
	ciphertext, err := base64.RawStdEncoding.DecodeString(encryptedMessage)
	if err != nil {
		return "", err
	}

	// Check if the ciphertext is too short
	nonceSize := keyCipher.NonceSize()
	if nonceSize == 0 || len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short or invalid nonce size")
	}

	// Extract the nonce and decrypt the ciphertext
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := keyCipher.Open(nil, nonce, ciphertext, []byte(associatedData))
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// encryptAndEncodeB64WithAES encrypts the plaintext using the AES key and encodes it as a base64 string.
func encryptAndEncodeB64WithAES(keyCipher cipher.AEAD, plaintext string) (string, error) {
	// Generate a random nonce
	nonce := make([]byte, keyCipher.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt the plaintext
	ciphertext := keyCipher.Seal(nonce, nonce, []byte(plaintext), []byte(associatedData))
	return base64.RawStdEncoding.EncodeToString(ciphertext), nil
}

// getenv returns the value of the environment variable with the given key.
func getenv[T any](key string, fallback T) (out T) {
	raw, ok := os.LookupEnv(key)
	if !ok || raw == "" {
		return fallback
	}

	target := reflect.ValueOf(&out).Elem()

	switch target.Kind() {
	case reflect.String:
		target.SetString(raw)
		return

	case reflect.Bool:
		parsed, err := strconv.ParseBool(raw)
		if err == nil {
			target.SetBool(parsed)
		}
		return

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err == nil {
			target.SetInt(parsed)
		}
		return

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		parsed, err := strconv.ParseUint(raw, 10, 64)
		if err == nil {
			target.SetUint(parsed)
		}
		return

	case reflect.Float32, reflect.Float64:
		parsed, err := strconv.ParseFloat(raw, 64)
		if err == nil {
			target.SetFloat(parsed)
		}
		return

	}

	return fallback
}

// parseAES2GCMKeyFromPem parses the AES key from the PEM-encoded data.
// It returns the AES GCM cipher or an error if parsing fails.
// The PEM-encoded data is expected to contain the AES key.
func parseAES2GCMKeyFromPem(raw []byte) (cipher.AEAD, error) {
	decoded, _ := pem.Decode(raw)
	if decoded == nil || decoded.Type != "AES PRIVATE KEY" {
		return nil, fmt.Errorf("failed to decode encryption key")
	}

	block, err := aes.NewCipher(decoded.Bytes)
	if err != nil {
		return nil, err
	}

	return cipher.NewGCM(block)
}

// parsePostgresURL parses the database URL and returns the DSN.
func parsePostgresURL(databaseUrl string) (string, error) {
	out := map[string]string{
		"host":     "localhost",
		"port":     "5432",
		"dbname":   "postgres",
		"user":     "postgres",
		"password": "pass",
		"sslmode":  "disable",
		"timezone": "Europe/Berlin",
	}

	// Parse the database URL
	databaseAddress, err := url.Parse(databaseUrl)
	if err != nil {
		return "", err
	}

	// Helper function to update the `out` map
	write := func(key string, value string, force bool) {
		if value != "" || force {
			out[key] = value
		}
	}

	// Write the host, port, and dbname
	write("host", databaseAddress.Hostname(), false)
	write("port", databaseAddress.Port(), false)
	write("dbname", strings.TrimPrefix(databaseAddress.Path, "/"), false)

	// Handle user credentials
	if databaseAddress.User != nil {
		write("user", databaseAddress.User.Username(), false)
		password, ok := databaseAddress.User.Password()
		write("password", password, ok)
	}

	// Handle query parameters (e.g., sslmode)
	for key, value := range databaseAddress.Query() {
		switch key {
		case "host", "port", "dbname", "user", "password":

		default:
			write(key, value[0], true)
		}
	}

	var parts []string
	for key, value := range out {
		parts = append(parts, fmt.Sprintf("%s=%s", key, value))
	}

	slices.Sort(parts)
	return strings.Join(parts, " "), nil
}

// selectValue returns the first non-zero value from the given list.
func selectValue[T comparable](values ...T) (zero T) {
	for _, value := range values {
		if value != zero {
			return value
		}
	}

	return
}
