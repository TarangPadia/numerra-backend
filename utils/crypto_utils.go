package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"strings"
)

func decodeAES256Key(keyB64 string) ([]byte, error) {
	rawKey, err := base64.RawURLEncoding.DecodeString(keyB64)
	if err != nil {
		return nil, err
	}
	if len(rawKey) != 32 {
		return nil, errors.New("invalid AES-256 key length (must be 32 bytes)")
	}
	return rawKey, nil
}

func Encrypt(plainText, keyB64 string) (string, error) {
	rawKey, err := decodeAES256Key(keyB64)
	if err != nil {
		return "", err
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

func Decrypt(encText, keyB64 string) (string, error) {
	rawKey, err := decodeAES256Key(keyB64)
	if err != nil {
		return "", err
	}

	data, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(encText))
	if err != nil {
		return "", err
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

func ParseDecryptedInvite(decrypted string) []string {
	return strings.Split(decrypted, "|")
}
