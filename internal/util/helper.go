package util

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateRandomID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
