package bot

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
)

var Token string
var dbv *sql.DB

var dancers = [...]string{"💃", "💃🏻", "💃🏼", "💃🏽", "💃🏾", "💃🏿", "🕺🏿", "🕺🏾", "🕺🏽", "🕺🏼", "🕺🏻", "🕺"}

type DiscordWebhookMessage struct {
	Content string `json:"content"`
}

func checkNilErr(e error) {
	if e != nil {
		log.Fatal("Error message")
	}
}

const pr = "%%"
const rankUsedEmojisInGuild = pr + "rankUsedEmojisInGuild"

func Run(dbc *sql.DB) {

	dbv = dbc

	// create a session
	discord, err := discordgo.New("Bot " + Token)
	checkNilErr(err)

	// add a event handler
	discord.AddHandler(newMessage)
	discord.AddHandler(messageUpdated)
	discord.AddHandler(newReaction)
	discord.AddHandler(removedReaction)
	discord.AddHandler(removedAllReactions)

	// open session
	err = discord.Open()

	NotifyAboutErrorViaWebhook(err)

	defer discord.Close() // close session, after function termination

	// keep bot running until there is NO os interruption (ctrl + C)
	fmt.Println("Bot running....")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

func newMessage(discord *discordgo.Session, message *discordgo.MessageCreate) {

	/* prevent bot responding to its own message
	this is achived by looking into the message author id
	if message.author.id is same as bot.author.id then just return
	*/
	if message.Author.ID == discord.State.User.ID || message.Author.Bot {
		return
	}

	// respond to user message if it contains `!help` or `!bye`
	switch {
	case strings.HasPrefix(message.Content, "%%hello"):
		discord.ChannelMessageSend(message.ChannelID, "Hello World😃")
	case strings.HasPrefix(message.Content, "%%bye"):
		discord.ChannelMessageSend(message.ChannelID, "Good Bye👋")

	case strings.HasPrefix(message.Content, "%%saveEverythingAboutThisGuild"):
		discord.ChannelMessageSend(message.ChannelID, "Ok")
		saveGuildInfo(discord, message.GuildID, dbv)
		discord.ChannelMessageSendReply(message.ChannelID, "Done", message.Reference())

	case strings.HasPrefix(message.Content, "%%danceInEveryChannel"):
		danceInEveryChannel(discord, message)

	case strings.HasPrefix(message.Content, "%%danceHere"):
		danceHere(discord, message)

	case strings.HasPrefix(message.Content, "%%helpMeRankEmojis"):
		discord.ChannelMessageSend(message.ChannelID, "This is an example, figure it out: %%rankUsedEmojisInGuild author=123 channel=123321 ignoreReactions=true belongToTheGuild=false ignoreMessageText=false fromDate=2022-01-01 toDate=2024-01-01 desc=true limit=10")
	case strings.HasPrefix(message.Content, "%%helpMeRankReactions"):
		discord.ChannelMessageSend(message.ChannelID, "This is a special case messageAuthor only works like this(dates are optional): %%rankUsedEmojisInGuild messageAuthor=123 ignoreMessageText=true fromDate=2022-01-01 toDate=2024-01-01 desc=true limit=10")
	case strings.HasPrefix(message.Content, rankUsedEmojisInGuild):
		s := ExtractSettings(message.Content)

		rankedEmojis, err := getRankedUsedEmojisInGuild(dbv, message.GuildID, s)

		if err != nil {
			NotifyAboutErrorViaWebhook(err)
			_, err = discord.ChannelMessageSend(message.ChannelID, "There was an error, try again later")
			return
		}

		embeds := rankedEmojis.TransformIntoDiscordEmbeds(s.ToJsonString())

		for _, embed := range embeds {
			_, err = discord.ChannelMessageSendEmbed(message.ChannelID, &embed)

			NotifyAboutErrorViaWebhook(err)
		}
	}

	err := ProcessOneMessage(nil, MessageModel{Message: message.Message}, message.GuildID, dbv, true)

	NotifyAboutErrorViaWebhook(err)

	return
}

func messageUpdated(_ *discordgo.Session, message *discordgo.MessageUpdate) {
	err := ProcessOneMessage(nil, MessageModel{Message: message.Message}, message.GuildID, dbv, true)

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

	err = ProcessOneMessage(discord, MessageModel{Message: msg}, messageReaction.GuildID, dbv, true)

	NotifyAboutErrorViaWebhook(err)
}

func NotifyAboutErrorViaWebhook(botErr error) {
	if botErr == nil {
		return
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
