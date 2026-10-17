package auth

import (
	"fmt"
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) (string, error) {
	keyData := headers.Get("Authorization")

	keyArr := strings.Fields(keyData)

	if len(keyArr) != 2 {
		return "", fmt.Errorf("Unable to extract key from %s", keyData)
	}

	return keyArr[1], nil
}
