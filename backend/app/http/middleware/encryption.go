package middleware

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
)

const (
	aesKey = "chloefuckityoall" // 16 bytes for AES-128
	aesIV  = "9311019310287172" // 16 bytes IV
)

// EncryptAES encrypts plaintext using AES-128-CBC
func EncryptAES(plaintext string) (string, error) {
	block, err := aes.NewCipher([]byte(aesKey))
	if err != nil {
		return "", err
	}

	// PKCS7 padding
	data := []byte(plaintext)
	padLen := aes.BlockSize - len(data)%aes.BlockSize
	padding := make([]byte, padLen)
	for i := range padding {
		padding[i] = byte(padLen)
	}
	data = append(data, padding...)

	// Encrypt
	ciphertext := make([]byte, len(data))
	mode := cipher.NewCBCEncrypter(block, []byte(aesIV))
	mode.CryptBlocks(ciphertext, data)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptAES decrypts AES-128-CBC encrypted text
func DecryptAES(encrypted string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher([]byte(aesKey))
	if err != nil {
		return "", err
	}

	if len(ciphertext)%aes.BlockSize != 0 {
		return "", fmt.Errorf("ciphertext is not a multiple of block size")
	}

	plaintext := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, []byte(aesIV))
	mode.CryptBlocks(plaintext, ciphertext)

	// Remove PKCS7 padding
	padLen := int(plaintext[len(plaintext)-1])
	if padLen > aes.BlockSize || padLen == 0 {
		return "", fmt.Errorf("invalid padding")
	}
	plaintext = plaintext[:len(plaintext)-padLen]

	return string(plaintext), nil
}

// EncryptionEnabled controls whether response encryption is active.
// Set to true to enable encryption for AI novel API responses.
var EncryptionEnabled = false
