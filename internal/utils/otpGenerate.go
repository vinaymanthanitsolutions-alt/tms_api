package utils

import (
	
	"crypto/rand"
	"fmt"
	"math/big"
	
)
func GenerateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(10000))
	if err != nil {
		return "", err
	}

	otp := fmt.Sprintf("%04d", n.Int64())
	return otp, nil
}
