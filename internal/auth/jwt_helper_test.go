package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestMakeJWTAndValidateJWT_Success(t *testing.T) {
	secret := "super-secret-key"
	userID := uuid.New()

	tokenString, err := MakeJWT(userID, secret, time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT returned unexpected error: %v", err)
	}

	gotUserID, err := ValidateJWT(tokenString, secret)
	if err != nil {
		t.Fatalf("ValidateJWT returned unexpected error: %v", err)
	}

	if gotUserID != userID {
		t.Fatalf("ValidateJWT() = %v, want %v", gotUserID, userID)
	}
}

func TestValidateJWT_RejectsWrongSecret(t *testing.T) {
	secret := "correct-secret"
	tokenString, err := MakeJWT(uuid.New(), secret, time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT returned unexpected error: %v", err)
	}

	if _, err := ValidateJWT(tokenString, "wrong-secret"); err == nil {
		t.Fatal("ValidateJWT() succeeded with an invalid secret, want error")
	}
}

func TestValidateJWT_RejectsMalformedToken(t *testing.T) {
	if _, err := ValidateJWT("not-a-jwt", "secret"); err == nil {
		t.Fatal("ValidateJWT() succeeded with malformed token, want error")
	}
}

func TestValidateJWT_RejectsExpiredToken(t *testing.T) {
	secret := "super-secret-key"
	userID := uuid.New()

	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   userID.String(),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		Issuer:    "chirpy-access",
	}).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("SignedString returned unexpected error: %v", err)
	}

	if _, err := ValidateJWT(tokenString, secret); err == nil {
		t.Fatal("ValidateJWT() succeeded with expired token, want error")
	}
}

func TestValidateJWT_RejectsNonUUIDSubject(t *testing.T) {
	secret := "super-secret-key"

	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   "not-a-uuid",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    "chirpy-access",
	}).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("SignedString returned unexpected error: %v", err)
	}

	if _, err := ValidateJWT(tokenString, secret); err == nil {
		t.Fatal("ValidateJWT() succeeded with non-UUID subject, want error")
	}
}
