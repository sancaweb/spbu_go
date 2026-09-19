package helper

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"os"
	"strings"
)

var ErrEncryptionSecretMissing = errors.New("SESSION_SECRET belum dikonfigurasi")

func encryptionKey() ([]byte, error) {
	secret := strings.TrimSpace(os.Getenv("SESSION_SECRET"))
	if secret == "" {
		return nil, ErrEncryptionSecretMissing
	}
	key := sha256.Sum256([]byte(secret))
	return key[:], nil
}

func EncryptSecret(value string) (string, error) {
	key, err := encryptionKey()
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(gcm.Seal(nonce, nonce, []byte(value), nil)), nil
}

func DecryptSecret(value string) (string, error) {
	key, err := encryptionKey()
	if err != nil {
		return "", err
	}
	data, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return "", errors.New("secret tersimpan tidak valid")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(data) < gcm.NonceSize() {
		return "", errors.New("secret tersimpan tidak valid")
	}
	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", errors.New("secret tersimpan tidak dapat dibuka")
	}
	return string(plaintext), nil
}
