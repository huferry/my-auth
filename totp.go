package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

func generateTOTP(secret []byte, t time.Time) string {
	// If the secret looks base32-encoded, decode it first
	if isBase32(secret) {
		decoded, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(string(secret)))
		if err == nil {
			secret = decoded
		}
	}

	counter := uint64(t.Unix()) / 30

	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, counter)

	mac := hmac.New(sha1.New, secret)
	mac.Write(buf)
	hash := mac.Sum(nil)

	offset := hash[len(hash)-1] & 0x0f
	code := binary.BigEndian.Uint32(hash[offset:offset+4]) & 0x7fffffff

	return fmt.Sprintf("%06d", code%1000000)
}

func isBase32(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	for _, b := range data {
		c := b
		if !(c >= 'A' && c <= 'Z') && !(c >= '2' && c <= '7') && c != '=' {
			return false
		}
	}
	return true
}
