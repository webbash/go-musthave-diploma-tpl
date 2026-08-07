package auth

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid token")

func GenerateJWT(subject string, secret string, ttl time.Duration) (string, error) {
	if secret == "" {
		return "", fmt.Errorf("empty secret: %w", ErrInvalidToken)
	}

	now := time.Now().UTC()
	claims := jwt.RegisteredClaims{
		Subject:   subject,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ParseJWT(token string, secret string) (string, error) {
	if token == "" || secret == "" {
		return "", ErrInvalidToken
	}

	claims := &jwt.RegisteredClaims{}
	parsed, err := jwt.ParseWithClaims(
		token,
		claims,
		func(t *jwt.Token) (any, error) {
			if t.Method != jwt.SigningMethodHS256 {
				return nil, ErrInvalidToken
			}
			return []byte(secret), nil
		},
		jwt.WithTimeFunc(time.Now),
	)
	if err != nil || !parsed.Valid || claims.Subject == "" {
		return "", ErrInvalidToken
	}

	return claims.Subject, nil
}

func SubjectToUserID(subject string) (int64, error) {
	id, err := strconv.ParseInt(subject, 10, 64)
	if err != nil {
		return 0, ErrInvalidToken
	}
	return id, nil
}
