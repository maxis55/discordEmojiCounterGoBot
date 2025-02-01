package bot

import (
	"bytes"
	"emoji-counter/db"
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"net/http"
	"os"
)

func messageUpdated(_ *discordgo.Session, message *discordgo.MessageUpdate) {
	err := ProcessOneMessage(nil, MessageModel{Message: message.Message}, message.GuildID, db.Connection, true)

	NotifyAboutErrorViaWebhook(err)
}

func newReaction(discord *discordgo.Session, messageReaction *discordgo.MessageReactionAdd) {
	processReaction(discord, messageReaction.MessageReaction)
}

func removedReaction(discord *discordgo.Session, messageReaction *discordgo.MessageReactionRemove) {
	processReaction(discord, messageReaction.MessageReaction)
}

func removedAllReactions(discord *discordgo.Session, messageReaction *discordgo.MessageReactionRemoveAll) {
	processReaction(discord, messageReaction.MessageReaction)
}

func processReaction(discord *discordgo.Session, messageReaction *discordgo.MessageReaction) {
	// ignore your own reactions just in case
	if messageReaction.UserID == discord.State.User.ID {
		return
	}

	//We don't know if it's a reaction under the bot message or not, reactions under the bot messages are ignored
	//so we need to re-fetch it
	msg, err := discord.ChannelMessage(messageReaction.ChannelID, messageReaction.MessageID, requestConfig)
	msg.GuildID = messageReaction.GuildID

	if err != nil {
		NotifyAboutErrorViaWebhook(err)
		return
	}

	err = ProcessOneMessage(discord, MessageModel{Message: msg}, messageReaction.GuildID, db.Connection, true)

	NotifyAboutErrorViaWebhook(err)
}

func rankingHandler(discord *discordgo.Session, message *discordgo.MessageCreate) {
	s := ExtractSettings(message.Content)

	rankedEmojis, err := getRankedUsedEmojisInGuild(db.Connection, message.GuildID, s)

	if err != nil {
		NotifyAboutErrorViaWebhook(err)
		return
	}

	embeds := rankedEmojis.TransformIntoDiscordEmbeds(s.ToJsonString())

	for _, embed := range embeds {
		_, err = discord.ChannelMessageSendEmbed(message.ChannelID, &embed)

		NotifyAboutErrorViaWebhook(err)
	}
}

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
