package emoji_processing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
)

func NotifyAboutErrorViaWebhook(botErr error) {
	if botErr == nil {
		return
	}

	// Always log to stdout first so the error survives even if Discord/the webhook is down.
	slog.Error("bot error", "err", botErr)

	webhookURL := os.Getenv("DISCORD_WEBHOOK_URL")
	if webhookURL == "" {
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
		slog.Warn("webhook marshal failed", "err", err)
		return
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		slog.Warn("webhook post failed", "err", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		slog.Warn("webhook unexpected response", "status", resp.Status)
	}
}
