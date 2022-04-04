package util

import (
	"crypto/sha256"
	"encoding/hex"
)

func ShaSum(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
