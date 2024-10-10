package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"sts-issuer/internal/sts"
)

// Define your defaults here
const (
	defaultAlias = "StsIssuer"
	defaultEmoji = ":key:"
	defaultColor = "#8fce00"
)

// Attachment represents the Rocket.Chat message attachment structure
type Attachment struct {
	Text  string `json:"text"`
	Color string `json:"color"`
}

// RocketChatPayload represents the complete payload structure for Rocket.Chat
type RocketChatPayload struct {
	Alias       string       `json:"alias"`
	Emoji       string       `json:"emoji"`
	Channel     string       `json:"channel"`
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments"`
}

// SendRocketChatNotification sends a custom message to Rocket.Chat using the webhook URL
func SendRocketChatNotification(stsCreds *sts.Creds, id string) error {
	// Load needed env vars
	webhookURL := os.Getenv("RC_WEBHOOK")
	if webhookURL == "" {
		log.Fatal("RC_WEBHOOK environment variable is not set")
	}

	channel := os.Getenv("RC_CHANNEL_" + id)
	if channel == "" {
		log.Fatal("RC_CHANNEL environment variable is not set")
	}

	title := os.Getenv("STS_TITLE_" + id)
	if title == "" {
		log.Fatal("STS_TITLE_ environment variable is not set")
	}

	text := fmt.Sprintf(
		"export AWS_ACCESS_KEY_ID=\"%s\";\n\nexport AWS_SECRET_ACCESS_KEY=\"%s\";\n\nexport AWS_SESSION_TOKEN=\"%s\";",
		stsCreds.AccessKeyID, stsCreds.SecretAccessKey, stsCreds.SessionToken)

	// Set defaults if parameters are empty
	payload := RocketChatPayload{
		Alias:   defaultAlias,
		Emoji:   defaultEmoji,
		Channel: channel,
		Text:    fmt.Sprintf(":key: S3 Credentials: %s\n:alarm_clock: Exipires: ~%s", title, stsCreds.Expiration),
		Attachments: []Attachment{
			{
				Text:  text,
				Color: defaultColor,
			},
		},
	}

	// Convert the payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Send the POST request
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response from Rocket.Chat: %s", resp.Status)
	}

	log.Println("RocketChat notification sent successfully")

	return nil
}
