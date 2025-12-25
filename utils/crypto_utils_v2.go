package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

func EncryptV2(plainText, keyB64 string) (string, error) {
	rawKey, err := base64.RawURLEncoding.DecodeString(keyB64)
	if err != nil {
		return "", err
	}
	if len(rawKey) != 32 {
		return "", errors.New("invalid AES-256 key length (must be 32 bytes)")
	}

	block, err := aes.NewCipher(rawKey)
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

	ciphertext := gcm.Seal(nonce, nonce, []byte(plainText), nil)
	return base64.RawURLEncoding.EncodeToString(ciphertext), nil
}

func DecryptV2(encText, keyB64 string) (string, error) {
	data, err := base64.RawURLEncoding.DecodeString(encText)
	if err != nil {
		return "", err
	}

	rawKey, err := base64.RawURLEncoding.DecodeString(keyB64)
	if err != nil {
		return "", err
	}
	if len(rawKey) != 32 {
		return "", errors.New("invalid AES-256 key length (must be 32 bytes)")
	}

	block, err := aes.NewCipher(rawKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	if len(data) < gcm.NonceSize() {
		return "", errors.New("ciphertext too short")
	}
	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
