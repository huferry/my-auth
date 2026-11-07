package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/pbkdf2"
)

const (
	pbkdf2Iterations = 100000
	keyLength        = 32
)

func generateKey() ([]byte, error) {
	key := make([]byte, keyLength)
	_, err := rand.Read(key)
	return key, err
}

// encryptWithKey encrypts data with AES-256-GCM using a raw key.
// Format: nonce (12 bytes) || ciphertext+tag
func encryptWithKey(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	result := make([]byte, 0, len(nonce)+len(ciphertext))
	result = append(result, nonce...)
	result = append(result, ciphertext...)
	return result, nil
}

// decryptWithKey decrypts data produced by encryptWithKey.
func decryptWithKey(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("data too short")
	}

	nonce := data[:nonceSize]
	ciphertext := data[nonceSize:]

	return gcm.Open(nil, nonce, ciphertext, nil)
}

// decryptSeed decrypts a TOTP seed from the CSV using AES-256-CBC with PBKDF2.
func decryptSeed(encryptedSeedB64, saltB64, ivHex, password string) ([]byte, error) {
	encryptedSeed, err := base64.StdEncoding.DecodeString(encryptedSeedB64)
	if err != nil {
		encryptedSeed, err = base64.RawStdEncoding.DecodeString(encryptedSeedB64)
		if err != nil {
			return nil, fmt.Errorf("decoding seed: %w", err)
		}
	}

	salt := []byte(saltB64)

	iv, err := hex.DecodeString(ivHex)
	if err != nil {
		return nil, fmt.Errorf("decoding iv: %w", err)
	}

	key := pbkdf2.Key([]byte(password), salt, pbkdf2Iterations, keyLength, sha1.New)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(encryptedSeed)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("ciphertext not a multiple of block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(encryptedSeed))
	mode.CryptBlocks(plaintext, encryptedSeed)

	// Remove PKCS7 padding
	if len(plaintext) == 0 {
		return nil, fmt.Errorf("empty plaintext")
	}
	padLen := int(plaintext[len(plaintext)-1])
	if padLen > aes.BlockSize || padLen == 0 {
		return nil, fmt.Errorf("invalid padding")
	}
	for i := 0; i < padLen; i++ {
		if plaintext[len(plaintext)-1-i] != byte(padLen) {
			return nil, fmt.Errorf("invalid padding")
		}
	}
	plaintext = plaintext[:len(plaintext)-padLen]

	return plaintext, nil
}
