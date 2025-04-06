package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type EmailRequest struct {
	To      string `json:"to"`
	From    string `json:"from"`
	Subject string `json:"subject"`
	Html    string `json:"html"`
}

func sendEmail(to, subject, html string) error {
	email := EmailRequest{
		To:      to,
		From:    os.Getenv("EMAIL_FROM"),
		Subject: subject,
		Html:    html,
	}

	body, err := json.Marshal(email)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("EMAIL_SERVER_PASSWORD")))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send email: %s", resp.Status)
	}

	return nil
}
