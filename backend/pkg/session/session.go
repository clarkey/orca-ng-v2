package session

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

const SessionTokenLength = 32

func GenerateToken() (string, error) {
	b := make([]byte, SessionTokenLength)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate session token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}