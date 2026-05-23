package bot

import (
	"emoji-counter/bot/emoji_processing"
	"emoji-counter/db"
	"log/slog"

	"github.com/bwmarrin/discordgo"
)

// rateLimited fires when discordgo hits a Discord rate limit and has to back
// off. It retries automatically; this just makes the event visible in logs so
// slow historical imports stop being a mystery.
func rateLimited(_ *discordgo.Session, rl *discordgo.RateLimit) {
	slog.Warn("discord rate limited",
		"url", rl.URL,
		"bucket", rl.Bucket,
		"retry_after", rl.RetryAfter,
		"message", rl.Message,
	)
}

func messageUpdated(_ *discordgo.Session, message *discordgo.MessageUpdate) {
	err := emoji_processing.ProcessOneMessage(nil, emoji_processing.MessageModel{Message: message.Message}, message.GuildID, db.Connection)

	emoji_processing.NotifyAboutErrorViaWebhook(err)
}

func messageDeleted(_ *discordgo.Session, message *discordgo.MessageDelete) {
	err := emoji_processing.ForgetEverythingAboutMessage(message.ID, db.Connection)

	emoji_processing.NotifyAboutErrorViaWebhook(err)
}

func newReaction(discord *discordgo.Session, messageReaction *discordgo.MessageReactionAdd) {
	processReaction(discord, messageReaction.MessageReaction, messageReaction.Member)
}

func removedReaction(discord *discordgo.Session, messageReaction *discordgo.MessageReactionRemove) {
	processReaction(discord, messageReaction.MessageReaction, nil)
}

func removedAllReactions(discord *discordgo.Session, messageReaction *discordgo.MessageReactionRemoveAll) {
	processReaction(discord, messageReaction.MessageReaction, nil)
}

func processReaction(discord *discordgo.Session, messageReaction *discordgo.MessageReaction, reactor *discordgo.Member) {
	// ignore your own reactions just in case
	if messageReaction.UserID == discord.State.User.ID {
		return
	}

	// Backfill the reactor as an author so emoji_used.author_id points at a real row.
	// Reactor may have never posted in the guild. On add events Member.User is set;
	// on remove events we only have the ID, so insert a minimal row.
	if reactor != nil && reactor.User != nil {
		_ = emoji_processing.AuthorModel{Author: reactor.User}.Remember(db.Connection)
	} else if messageReaction.UserID != "" {
		_ = emoji_processing.RememberAuthorByID(messageReaction.UserID, db.Connection)
	}

	//We don't know if it's a reaction under the bot message or not, reactions under the bot messages are ignored
	//so we need to re-fetch it
	msg, err := discord.ChannelMessage(messageReaction.ChannelID, messageReaction.MessageID, emoji_processing.RequestConfig)
	if err != nil {
		emoji_processing.NotifyAboutErrorViaWebhook(err)
		return
	}
	if msg == nil {
		// Message was deleted between the reaction event and the fetch.
		return
	}
	msg.GuildID = messageReaction.GuildID

	err = emoji_processing.ProcessOneMessage(discord, emoji_processing.MessageModel{Message: msg}, messageReaction.GuildID, db.Connection)

	emoji_processing.NotifyAboutErrorViaWebhook(err)
}

func rankingHandler(discord *discordgo.Session, message *discordgo.MessageCreate) {
	s := emoji_processing.ExtractSettings(message.Content)

	rankedEmojis, err := emoji_processing.GetRankedUsedEmojisInGuild(db.Connection, message.GuildID, s)

	if err != nil {
		emoji_processing.NotifyAboutErrorViaWebhook(err)
		return
	}

	embeds := rankedEmojis.TransformIntoDiscordEmbeds(s.ToJsonString())

	for _, embed := range embeds {
		_, err = discord.ChannelMessageSendEmbed(message.ChannelID, &embed)

		emoji_processing.NotifyAboutErrorViaWebhook(err)
	}
}

func handleHistoricalForGuild(d *discordgo.Session, message *discordgo.MessageCreate) {
	channels, err := emoji_processing.QueryAllGuildChannels(db.Connection, message.GuildID)
	if err != nil {
		d.ChannelMessageSendReply(message.ChannelID, "💀 Reason: "+err.Error(), message.Reference(), emoji_processing.RequestConfig)
		return
	}

	emoji_processing.Dance(d, message, channels)
}

func handleHistoricalForChannel(d *discordgo.Session, message *discordgo.MessageCreate) {
	var channels []emoji_processing.ChannelModel
	channel, err := emoji_processing.QueryChannelById(db.Connection, message.ChannelID)

	if err != nil {
		d.ChannelMessageSendReply(message.ChannelID, "💀 Reason: "+err.Error(), message.Reference(), emoji_processing.RequestConfig)
		return
	}

	if channel == nil {
		d.ChannelMessageSendReply(message.ChannelID, "Cant find the channel in the DB. Save this guild first maybe", message.Reference(), emoji_processing.RequestConfig)
		return
	}

	channels = append(channels, *channel)

	emoji_processing.Dance(d, message, channels)
}
