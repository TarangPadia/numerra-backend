package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"log"
	"strings"
)

func Encrypt(plainText, key string) (string, error) {
	rawKey, err := base64.RawURLEncoding.DecodeString(key)
	if err != nil {
		log.Fatalf("Failed to decode Base64 key: %v", err)
	}
	if len(rawKey) != 32 {
		log.Fatalf("Expected 32 bytes for AES-256 key, got %d", len(rawKey))
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

	ciphertext := gcm.Seal(nonce, nonce, []byte(plainText), nil)

	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

func Decrypt(encText, key, decryptType string) (string, error) {
	var data []byte
	var err error
	if decryptType == "INVITE" {
		data, err = base64.URLEncoding.DecodeString(encText)
	} else {
		data, err = base64.StdEncoding.DecodeString(encText)
	}
	if err != nil {
		return "", err
	}

	rawKey, err := base64.RawURLEncoding.DecodeString(key)
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

func ParseDecryptedInvite(decrypted string) []string {
	return strings.Split(decrypted, "|")
}
