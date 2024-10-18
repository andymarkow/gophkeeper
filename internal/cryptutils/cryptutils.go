// Package cryptutils provides cryptography utils.
package cryptutils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	"golang.org/x/crypto/scrypt"
)

// EncryptString encrypts a string using AES in CFB mode. The key should be kept
// secret. The resulting ciphertext is base64 encoded and can be safely stored
// in a database or file. The function always returns a string of the same
// length, so it can be used without revealing the length of the plaintext.
func EncryptString(key []byte, plaintext string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("aes.NewCipher: %w", err)
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))

	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", fmt.Errorf("io.ReadFull: %w", err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], []byte(plaintext))

	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

// DecryptString decrypts a string using AES in CFB mode. The key should be kept
// secret. The function will return an error if the ciphertext is invalid or if
// the key is invalid.
func DecryptString(key []byte, ciphertext string) (string, error) {
	ciphertextBytes, err := base64.URLEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("base64.DecodeString: %w", err)
	}

	if len(ciphertextBytes) < aes.BlockSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("aes.NewCipher: %w", err)
	}

	iv := ciphertextBytes[:aes.BlockSize]
	ciphertextBytes = ciphertextBytes[aes.BlockSize:]

	textBytes := make([]byte, len(ciphertextBytes))

	mode := cipher.NewCFBDecrypter(block, iv)
	mode.XORKeyStream(textBytes, ciphertextBytes)

	return string(textBytes), nil
}

func EncryptStream(passphrase string, rd io.Reader) (io.Reader, error) {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, fmt.Errorf("rand.Read: %w", err)
	}

	key, err := scrypt.Key([]byte(passphrase), salt, 1<<15, 8, 1, 32)
	if err != nil {
		return nil, fmt.Errorf("scrypt.Key: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes.NewCipher: %w", err)
	}

	// gcm, err := cipher.NewGCM(block)
	// if err != nil {
	// 	return nil, fmt.Errorf("cipher.NewGCM: %w", err)
	// }

	nonce := make([]byte, block.BlockSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, fmt.Errorf("io.ReadFull: %w", err)
	}

	stream := cipher.NewCTR(block, nonce)

	// Create encrypted stream reader.
	encryptReader := cipher.StreamReader{
		S: stream,
		R: io.MultiReader(bytes.NewReader(salt), bytes.NewReader(nonce), rd),
	}

	return encryptReader, nil
}
