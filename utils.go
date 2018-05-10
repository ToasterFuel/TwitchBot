package main

import "crypto/hmac"
import "crypto/sha256"
import "encoding/hex"
import "github.com/nu7hatch/gouuid"
import "math/rand"

func checkSignature(message []byte, expectedSignature string, key []byte) bool {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	computedSignature := hex.EncodeToString(mac.Sum(nil))

	return expectedSignature == computedSignature
}

func getRequestId() string {
	uuid, err := uuid.NewV4()
	if err != nil {
		return string(rand.Intn(100000))
	}

	return uuid.String()
}
