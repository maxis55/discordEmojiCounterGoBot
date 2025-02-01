package emoji_processing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

func NotifyAboutErrorViaWebhook(botErr error) {
	if botErr == nil {
		return
	}

	type DiscordWebhookMessage struct {
		Content string `json:"content"`
	}

	payload := DiscordWebhookMessage{
		Content: fmt.Sprintf("<@%s> 💀 Reason: %s", os.Getenv("BOT_MASTER_ID"), botErr.Error()),
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Println(fmt.Sprintf("Failed to marshal payload: %v", err))
		return
	}

	resp, err := http.Post(os.Getenv("DISCORD_WEBHOOK_URL"), "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Println(fmt.Sprintf("Failed to send webhook: %v", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		log.Println(fmt.Sprintf("Unexpected response from Discord: %s", resp.Status))
		return
	}
}
