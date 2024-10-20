// Package cryptutils provides cryptography utils.
package cryptutils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash"
	"hash/crc32"
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

// generateAESKey generates a random salt and then uses the passphrase to generate an
// AES key using the scrypt key derivation function. The key is 32 bytes long.
//
// The generated salt is also returned, and should be stored securely along
// with the encrypted data. The salt should not be secret.
func generateAESKey(passphrase []byte) ([]byte, []byte, error) {
	// Generate random salt with the size of 16 bytes.
	salt := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, salt)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	// Generate key based on passphrase and salt.
	key, err := scrypt.Key(passphrase, salt, 1<<15, 8, 1, 32)
	if err != nil {
		return nil, nil, fmt.Errorf("scrypt.Key: %w", err)
	}

	return key, salt, nil
}

type EncryptedStream struct {
	salt   []byte
	iv     []byte
	stream io.Reader
}

func (s *EncryptedStream) Salt() []byte {
	return s.salt
}

func (s *EncryptedStream) SaltHex() string {
	return hex.EncodeToString(s.salt)
}

func (s *EncryptedStream) IV() []byte {
	return s.iv
}

func (s *EncryptedStream) IVHex() string {
	return hex.EncodeToString(s.iv)
}

func (s *EncryptedStream) Stream() io.Reader {
	return s.stream
}

func EncryptStream(passphrase []byte, rd io.Reader) (*EncryptedStream, error) {
	key, salt, err := generateAESKey(passphrase)
	if err != nil {
		return nil, fmt.Errorf("failed to generate AES key: %w", err)
	}

	// Create cipher block based on the key.
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes.NewCipher: %w", err)
	}

	// Create Initialization Vector with the same size as the block size.
	iv := make([]byte, block.BlockSize())
	_, err = io.ReadFull(rand.Reader, iv)
	if err != nil {
		return nil, fmt.Errorf("io.ReadFull: %w", err)
	}

	// Create encryptor in CTR mode.
	stream := cipher.NewCTR(block, iv)

	// Create encrypted stream reader.
	encryptReader := cipher.StreamReader{
		S: stream,
		R: rd,
	}

	return &EncryptedStream{
		salt:   salt,
		iv:     iv,
		stream: encryptReader,
	}, nil
}

type readCloser struct {
	io.Reader
	io.Closer
}

func DecryptStream(passphrase, salt, iv []byte, rd io.ReadCloser) (io.ReadCloser, error) {
	// Generate key from passphrase and salt.
	key, err := scrypt.Key(passphrase, salt, 1<<15, 8, 1, 32)
	if err != nil {
		return nil, fmt.Errorf("scrypt.Key: %w", err)
	}

	// Create cipher block based on the key.
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes.NewCipher: %w", err)
	}

	// Create decryptor in CTR mode.
	stream := cipher.NewCTR(block, iv)

	// Create decrypted stream reader.
	decryptReader := cipher.StreamReader{
		S: stream,
		R: rd,
	}

	return &readCloser{
		Reader: decryptReader,
		Closer: rd,
	}, nil
}

// CalcStreamHash calculates the CRC32 hashsum of a given io.Reader.
func CalcStreamHash(rd io.Reader) (io.Reader, hash.Hash32) {
	// Create a CRC32 hash table.
	table := crc32.MakeTable(crc32.Castagnoli)

	// Create a CRC32 hash function using the table.
	hash := crc32.New(table)

	// Wrap the reader with the hash function writer.
	teeReader := io.TeeReader(rd, hash)

	return teeReader, hash
}
