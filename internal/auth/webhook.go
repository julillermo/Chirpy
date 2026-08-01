package auth

import (
	"errors"
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) (string, error) {
	authorizationHeader := headers.Get("Authorization")
	if len(authorizationHeader) <= 0 {
		return "", errors.New("The Authorization header doesn't exist")
	}

	headerParts := strings.Split(authorizationHeader, " ")

	return headerParts[1], nil
}
