package utils

import (
	"fmt"
	"log"
	"os"

	mail "gopkg.in/mail.v2"
)

func SendEmailGomail(to, subject, body string) error {
	m := mail.NewMessage()
	m.SetHeader("From", os.Getenv("SMTP_USER"))
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := mail.NewDialer(os.Getenv("SMTP_HOST"), parsePort(), os.Getenv("SMTP_USER"), os.Getenv("SMTP_PASSWORD"))
	d.StartTLSPolicy = mail.MandatoryStartTLS

	if err := d.DialAndSend(m); err != nil {
		log.Println("Failed to send email:", err)
		return err
	}
	return nil
}

func parsePort() int {
	portStr := os.Getenv("SMTP_PORT")
	if portStr == "" {
		return 587
	}
	var port int
	fmt.Sscanf(portStr, "%d", &port)
	return port
}
