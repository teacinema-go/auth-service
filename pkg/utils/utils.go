package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"strconv"
)

func GenerateCode() (int64, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(900_000))
	if err != nil {
		return 0, err
	}
	return 100_000 + n.Int64(), nil
}

func GenerateHashForCode(code int64) string {
	return GenerateHash([]byte(strconv.FormatInt(code, 10)))
}

func GenerateHash(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
