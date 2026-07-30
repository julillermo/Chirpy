package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func MakeJWT(userId uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	mySigningKey := []byte(tokenSecret)

	claims := &jwt.RegisteredClaims{
		Issuer:    "Chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:   userId.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(mySigningKey)
	return ss, err
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := &jwt.RegisteredClaims{}

	parser := jwt.NewParser()
	token, err := parser.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			return []byte(tokenSecret), nil
		},
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("Error with validating JWT: %v", err)
	}

	subjectId, subjectIdErr := token.Claims.GetSubject()
	if subjectIdErr != nil {
		fmt.Println("Error with subjectId from token: ", subjectIdErr)
		return uuid.Nil, subjectIdErr
	}

	subjectUUID, subjectUUIDErr := uuid.Parse(subjectId)
	if subjectUUIDErr != nil {
		fmt.Println("Error with converting subjectId to uuid.UUID: ", subjectUUIDErr)
		return uuid.Nil, subjectUUIDErr
	}

	return subjectUUID, nil
}

func GetBearerToken(header http.Header) (string, error) {
	authorizationHeader := header.Get("Authorization")
	if len(authorizationHeader) <= 0 {
		return "", errors.New("The Authorization header doesn't exist")
	}

	headerParts := strings.Split(authorizationHeader, " ")

	return headerParts[1], nil
}
