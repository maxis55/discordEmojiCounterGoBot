package bot

import (
	"database/sql"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"math/rand"
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

func Run(dbc *sql.DB) {

	dbv = dbc

	// create a session
	discord, err := discordgo.New("Bot " + Token)
	checkNilErr(err)

	// add a event handler
	discord.AddHandler(newMessage)
	discord.AddHandler(newReaction)

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
	case strings.Contains(message.Content, "%%hello"):
		discord.ChannelMessageSend(message.ChannelID, "Hello World😃")
	case strings.Contains(message.Content, "%%bye"):
		discord.ChannelMessageSend(message.ChannelID, "Good Bye👋")
	case strings.Contains(message.Content, "%%test"):
		discord.ChannelMessageSendReply(message.ChannelID, "Good Bye👋", message.Reference())
	case strings.Contains(message.Content, "%%bye"):
		discord.ChannelMessageSend(message.ChannelID, "Good Bye👋")
	case strings.Contains(message.Content, "%%saveEverythingAboutThisGuild"):
		discord.ChannelMessageSend(message.ChannelID, "Ok")
		saveGuildInfo(discord, message.GuildID, dbv)

	case strings.Contains(message.Content, "%%channelTest"):
		discord.ChannelMessageSend(message.ChannelID, "Ok"+dancers[rand.Intn(len(dancers))])

	case strings.Contains(message.Content, "%%danceInEveryChannel"):
		discord.ChannelMessageSend(message.ChannelID, "Ok"+dancers[rand.Intn(len(dancers))])

		channels, err := queryAllGuildChannels(dbv, message.GuildID)
		if err != nil {
			discord.ChannelMessageSendReply(message.ChannelID, "💀 Reason: "+err.Error(), message.Reference())
			return
		}

		for _, channel := range channels {
			go getAndSaveAllMessages(discord, message.Message, message.GuildID, dbv, make(map[string]AuthorModel), message.Reference(), channel.Channel)
		}

	case strings.Contains(message.Content, "%%dance"):
		discord.ChannelMessageSend(message.ChannelID, "Ok"+dancers[rand.Intn(len(dancers))])

		channel, err := queryChannelById(dbv, message.ChannelID)

		if err != nil {
			discord.ChannelMessageSendReply(message.ChannelID, "💀 Reason: "+err.Error(), message.Reference())
			return
		}

		if channel == nil {
			discord.ChannelMessageSendReply(message.ChannelID, "Cant find the channel in the DB. Save this guild first maybe", message.Reference())
			return
		}

		go getAndSaveAllMessages(discord, message.Message, message.GuildID, dbv, make(map[string]AuthorModel), message.Reference(), channel.Channel)

	}

	if message.Author.ID == "181180158441422848" {
		ProcessOneMessage(MessageModel{Message: message.Message}, message.GuildID, dbv, true)
	}
}

func newReaction(discord *discordgo.Session, messageReaction *discordgo.MessageReactionAdd) {

	///* prevent bot responding to its own message
	//this is achived by looking into the message author id
	//if message.author.id is same as bot.author.id then just return
	//*/
	//if message.UserID.ID == discord.State.User.ID || message.Author.Bot {
	//	return
	//}

	fmt.Println(fmt.Sprintf("%+v", messageReaction.Emoji))

}
