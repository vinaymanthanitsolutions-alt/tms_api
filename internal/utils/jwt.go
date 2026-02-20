package utils

import (
	// "backend/internal/utils"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GetJwtSecret() []byte {
    return []byte(GetEnv("SECRET_KEY"))
}

func GenerateToken(empID string, role string) (string, error) {
	claims := jwt.MapClaims{
		"emp_id": empID,
		"role":  role,
		"exp":    time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(GetJwtSecret())
}
