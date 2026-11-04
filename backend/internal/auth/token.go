package auth

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func CreateToken(userID int64, role string) (string, error) {

	secret := os.Getenv("JWT_SECRET")

	if len(secret) < 32 {
		return "", errors.New("JWT_SECRET is missing or too short")
	}

	claims := jwt.MapClaims{
		"sub":  strconv.FormatInt(userID, 10),
		"role": role,
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}
