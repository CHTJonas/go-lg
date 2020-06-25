package storage

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateUID() string {
	token := make([]byte, 6)
	rand.Read(token)
	return hex.EncodeToString(token)
}
