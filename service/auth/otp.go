package auth

import (
	"fmt"
	"math/rand"
	"time"
)

func GenerateOTP() string {
	rand.Seed(time.Now().UnixNano())
	otp := rand.Intn(899999) + 100000
	return fmt.Sprintf("%06d", otp)
}