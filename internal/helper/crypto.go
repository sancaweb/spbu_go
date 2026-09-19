package helper

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

var defaultSecretKey = []byte("spbu_go_secret_key_32_bytes_safe!") // 32 bytes for AES-256

// EncodePenyusutanToken mengenkripsi bbmID, shiftID, dan penjualanID menjadi URL-safe token.
func EncodePenyusutanToken(bbmID uint, shiftID uint, penjualanID uint64) string {
	plainText := fmt.Sprintf("%d:%d:%d", bbmID, shiftID, penjualanID)

	block, err := aes.NewCipher(defaultSecretKey[:32])
	if err != nil {
		return ""
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return ""
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return ""
	}

	cipherText := gcm.Seal(nonce, nonce, []byte(plainText), nil)
	return base64.RawURLEncoding.EncodeToString(cipherText)
}

// DecodePenyusutanToken mendeskripsi token kembali menjadi bbmID, shiftID, dan penjualanID.
func DecodePenyusutanToken(token string) (bbmID uint, shiftID uint, penjualanID uint64, err error) {
	data, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return 0, 0, 0, errors.New("token tidak valid")
	}

	block, err := aes.NewCipher(defaultSecretKey[:32])
	if err != nil {
		return 0, 0, 0, errors.New("gagal inisialisasi cipher")
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return 0, 0, 0, errors.New("gagal inisialisasi GCM")
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return 0, 0, 0, errors.New("format token tidak valid")
	}

	nonce, cipherText := data[:nonceSize], data[nonceSize:]
	plainTextBytes, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return 0, 0, 0, errors.New("dekripsi token gagal")
	}

	parts := strings.Split(string(plainTextBytes), ":")
	if len(parts) != 3 {
		return 0, 0, 0, errors.New("struktur payload token tidak valid")
	}

	bID, err1 := strconv.ParseUint(parts[0], 10, 32)
	sID, err2 := strconv.ParseUint(parts[1], 10, 32)
	pID, err3 := strconv.ParseUint(parts[2], 10, 64)
	if err1 != nil || err2 != nil || err3 != nil {
		return 0, 0, 0, errors.New("parsing ID gagal")
	}

	return uint(bID), uint(sID), pID, nil
}
