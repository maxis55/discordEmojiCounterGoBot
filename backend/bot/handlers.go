package bot

import (
	"emoji-counter/bot/emoji_processing"
	"emoji-counter/db"
	"github.com/bwmarrin/discordgo"
)

func messageUpdated(_ *discordgo.Session, message *discordgo.MessageUpdate) {
	err := emoji_processing.ProcessOneMessage(nil, emoji_processing.MessageModel{Message: message.Message}, message.GuildID, db.Connection)

	emoji_processing.NotifyAboutErrorViaWebhook(err)
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
	msg, err := discord.ChannelMessage(messageReaction.ChannelID, messageReaction.MessageID, emoji_processing.RequestConfig)
	msg.GuildID = messageReaction.GuildID

	if err != nil {
		emoji_processing.NotifyAboutErrorViaWebhook(err)
		return
	}

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
