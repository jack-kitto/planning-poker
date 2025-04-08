package utils

import (
	"crypto/rand"
	"encoding/base64"
	"log"
)

func GenerateToken() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	return base64.URLEncoding.EncodeToString(b)
}

func SendEmail(email, subject, body string) error {
	// Implement your email sending logic here
	log.Printf("Sending email to %s with subject %s", email, subject)
	return nil
}
