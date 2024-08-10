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
	"sync"
	"sync/atomic"
	"time"
)

var Token string
var dbv *sql.DB

var dancers = [...]string{"💃", "💃🏻", "💃🏼", "💃🏽", "💃🏾", "💃🏿", "🕺🏿", "🕺🏾", "🕺🏽", "🕺🏼", "🕺🏻", "🕺"}

type WaitGroupCount struct {
	sync.WaitGroup
	count int64
}

func (wg *WaitGroupCount) Add(delta int) {
	atomic.AddInt64(&wg.count, int64(delta))
	wg.WaitGroup.Add(delta)
}

func (wg *WaitGroupCount) Done() {
	atomic.AddInt64(&wg.count, -1)
	wg.WaitGroup.Done()
}

func (wg *WaitGroupCount) GetCount() int {
	return int(atomic.LoadInt64(&wg.count))
}

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
	case strings.Contains(message.Content, "%%help"):
		discord.ChannelMessageSend(message.ChannelID, "Nobody will help you")

	case strings.Contains(message.Content, "%%saveEverythingAboutThisGuild"):
		discord.ChannelMessageSend(message.ChannelID, "Ok")
		saveGuildInfo(discord, message.GuildID, dbv)
		discord.ChannelMessageSendReply(message.ChannelID, "Done", message.Reference())

	case strings.Contains(message.Content, "%%danceInEveryChannel"):
		trackingMsg, _ := discord.ChannelMessageSend(message.ChannelID, "Ok"+dancers[rand.Intn(len(dancers))])

		channels, err := queryAllGuildChannels(dbv, message.GuildID)
		if err != nil {
			discord.ChannelMessageSendReply(message.ChannelID, "💀 Reason: "+err.Error(), message.Reference(), requestConfig)
			return
		}

		wg := WaitGroupCount{}

		for _, channel := range channels {
			wg.Add(1)
			go func(ch ChannelModel) {
				defer wg.Done()
				getAndSaveAllMessages(discord, message.Message, message.GuildID, dbv, make(map[string]AuthorModel), message.Reference(), ch.Channel)
			}(channel)
		}
		for wg.GetCount() > 0 {
			trackingMsg, err = discord.ChannelMessageEdit(trackingMsg.ChannelID, trackingMsg.ID, fmt.Sprintf("Working on %d channels", wg.GetCount()), requestConfig)
			time.Sleep(time.Second * 2)
		}
		discord.ChannelMessageEdit(trackingMsg.ChannelID, trackingMsg.ID, "Done", requestConfig)

	case strings.Contains(message.Content, "%%danceHere"):
		trackingMsg, _ := discord.ChannelMessageSend(message.ChannelID, "Ok"+dancers[rand.Intn(len(dancers))])

		channel, err := queryChannelById(dbv, message.ChannelID)

		if err != nil {
			discord.ChannelMessageSendReply(message.ChannelID, "💀 Reason: "+err.Error(), message.Reference(), requestConfig)
			return
		}

		if channel == nil {
			discord.ChannelMessageSendReply(message.ChannelID, "Cant find the channel in the DB. Save this guild first maybe", message.Reference(), requestConfig)
			return
		}
		wg := WaitGroupCount{}
		wg.Add(1)
		go func(ch ChannelModel) {
			defer wg.Done()
			getAndSaveAllMessages(discord, message.Message, message.GuildID, dbv, make(map[string]AuthorModel), message.Reference(), ch.Channel)
		}(*channel)

		for wg.GetCount() > 0 {
			trackingMsg, err = discord.ChannelMessageEdit(trackingMsg.ChannelID, trackingMsg.ID, fmt.Sprintf("Working on %d channels", wg.GetCount()), requestConfig)
			time.Sleep(time.Second * 2)
		}
		discord.ChannelMessageEdit(trackingMsg.ChannelID, trackingMsg.ID, "Done", requestConfig)

	case strings.Contains(message.Content, "%%rankUsedEmojisInGuild"):
		res, err := rankUsedEmojisInGuild(dbv, message.GuildID, 10)

		if err != nil {
			discord.ChannelMessageSendReply(message.ChannelID, "💀 Reason: "+err.Error(), message.Reference(), requestConfig)
			return
		}

		discord.ChannelMessageSendReply(message.ChannelID, res, message.Reference())
	}

	err := ProcessOneMessage(MessageModel{Message: message.Message}, message.GuildID, dbv, true)

	if err != nil {
		discord.ChannelMessageSendReply(message.ChannelID, "💀 Reason: "+err.Error(), message.Reference(), requestConfig)
	}

	return

}

func newReaction(discord *discordgo.Session, messageReaction *discordgo.MessageReactionAdd) {

	///* prevent bot responding to its own message
	//this is achived by looking into the message author id
	//if message.author.id is same as bot.author.id then just return
	//*/
	if messageReaction.UserID == discord.State.User.ID {
		return
	}

	fmt.Println(fmt.Sprintf("%+v", messageReaction.Emoji))

}
