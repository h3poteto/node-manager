package requestid

import (
	"crypto/rand"
	"encoding/hex"
)

func RequestID() (string, error) {
	r := make([]byte, 12)
	_, err := rand.Read(r)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(r), nil
}
