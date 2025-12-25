package utils

import (
	"crypto/rand"
	"fmt"
)

func Generate6DigitOTP() string {
	b := make([]byte, 3)
	rand.Read(b)
	return fmt.Sprintf("%06d", int(b[0])%10*100000+int(b[1])%1000+int(b[2])%10)
}
