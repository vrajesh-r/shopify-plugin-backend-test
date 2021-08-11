package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"sort"
)

func GenerateGatewaySignature(values map[string]string, secret string) string {
	var fields sort.StringSlice
	for key, _ := range values {
		fields = append(fields, key)
	}
	fields.Sort()

	var concatenation string
	for _, key := range fields {
		concatenation += key + values[key]
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(concatenation))

	return hex.EncodeToString(mac.Sum(nil))
}
