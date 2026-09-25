package auth

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const authHeader = "Authorization"
const bearerString = "Bearer"

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:   userID.String(),
	})

	return token.SignedString([]byte(tokenSecret))
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	})

	if err != nil {
		return uuid.Nil, fmt.Errorf("unable to parse token string: %w", err)
	}

	subject, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, fmt.Errorf("unable to extract the subject: %w", err)
	}

	u, err := uuid.Parse(subject)
	if err != nil {
		return uuid.Nil, fmt.Errorf("unable to parse string into uuid: %w", err)
	}

	return u, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	bearerAuthString := headers.Get(authHeader)

	bearerAuthList := strings.Fields(bearerAuthString)

	if len(bearerAuthList) < 2 {
		return "", fmt.Errorf("Incorrect bearer token %s", bearerAuthString)
	}

	if strings.Compare(bearerAuthList[0], bearerString) != 0 {
		return "", fmt.Errorf("Incorrect bearer name %s", bearerAuthString)
	}

	return bearerAuthList[1], nil
}
