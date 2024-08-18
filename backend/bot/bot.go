package bot

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"os"
	"os/signal"
	"strings"
)

var Token string
var dbv *sql.DB

var dancers = [...]string{"💃", "💃🏻", "💃🏼", "💃🏽", "💃🏾", "💃🏿", "🕺🏿", "🕺🏾", "🕺🏽", "🕺🏼", "🕺🏻", "🕺"}

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
	discord.Open()
	defer discord.Close() // close session, after function termination

	// keep bot running untill there is NO os interruption (ctrl + C)
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
		discord.ChannelMessageSend(message.ChannelID, "This is an example, figure it out: %%rankUsedEmojisInGuild author=123 channel=123321 ignoreReactions=true belongToTheGuild=false ignoreMessageText=false fromDate=2022-01-01 toDate=2024-01-01 limit=10")
	case strings.HasPrefix(message.Content, "%%helpMeRankReactions"):
		discord.ChannelMessageSend(message.ChannelID, "This is a special case messageAuthor only works like this(dates are optional): %%rankUsedEmojisInGuild messageAuthor=123 ignoreMessageText=true fromDate=2022-01-01 toDate=2024-01-01 limit=10")
	case strings.HasPrefix(message.Content, rankUsedEmojisInGuild):
		s := ExtractSettings(message.Content)

		js, _ := json.Marshal(s)

		discord.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Rank with following settings: %s", js))

		res, err := getRankedUsedEmojisInGuild(dbv, message.GuildID, s)

		if err != nil {
			discord.ChannelMessageSendReply(message.ChannelID, "💀 Reason: "+err.Error(), message.Reference(), requestConfig)
			return
		}

		discord.ChannelMessageSendReply(message.ChannelID, res, message.Reference())
	}

	err := ProcessOneMessage(nil, MessageModel{Message: message.Message}, message.GuildID, dbv, true)

	if err != nil {
		discord.ChannelMessageSendReply(message.ChannelID, "💀 Reason: "+err.Error(), message.Reference(), requestConfig)
	}

	return
}

func messageUpdated(discord *discordgo.Session, message *discordgo.MessageUpdate) {
	err := ProcessOneMessage(nil, MessageModel{Message: message.Message}, message.GuildID, dbv, true)

	if err != nil {
		discord.ChannelMessageSendReply(message.ChannelID, "💀 Reason: "+err.Error(), message.Reference(), requestConfig)
	}
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

	//We don't know if its a reaction under the bot message or not, reactions under the bot messages are ignored
	//so we need to refetch it

	msg, err := discord.ChannelMessage(messageReaction.ChannelID, messageReaction.MessageID, requestConfig)
	msg.GuildID = messageReaction.GuildID

	if err != nil {
		discord.ChannelMessageSend(messageReaction.ChannelID, "💀 Reason: "+err.Error(), requestConfig)
	}

	err = ProcessOneMessage(discord, MessageModel{Message: msg}, messageReaction.GuildID, dbv, true)

	if err != nil {
		discord.ChannelMessageSend(messageReaction.ChannelID, "💀 Reason: "+err.Error(), requestConfig)
	}
}
