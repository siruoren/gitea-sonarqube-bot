package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
)

func isValidWebhook(message []byte, key string, signature string, component string) (bool, error) {
	log.Printf("'%s'", signature)

	if key == "" && signature == "" {
		// No webhook token configured and no signature header received. Skipping request validation.
		return true, nil
	}

	if key == "" && signature != "" {
		return false, fmt.Errorf("Signature header received but no %s webhook secret configured. Request rejected due to possible configuration mismatch.", component)
	}

	if key != "" && signature == "" {
		return false, fmt.Errorf("%s webhook secret configured but no signature header received. Request rejected due to possible configuration mismatch.", component)
	}

	decodedSignature, err := hex.DecodeString(signature)
	if err != nil {
		return false, fmt.Errorf("Error decoding signature for %s webhook.", component)
	}

	mac := hmac.New(sha256.New, []byte(key))
	mac.Write(message)
	sum := mac.Sum(nil)

	if !hmac.Equal(decodedSignature, sum) {
		return false, fmt.Errorf("Signature header does not match the received %s webhook content. Request rejected.", component)
	}

	return true, nil
}
