package bot

import (
	"emoji-counter/db"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"os"
	"os/signal"
	"strings"
)

func checkNilErr(e error) {
	if e != nil {
		log.Fatal("Error message")
	}
}

const pr = "%%"
const rankUsedEmojisInGuild = pr + "rankUsedEmojisInGuild"

func Run() {

	// create a session
	discord, err := discordgo.New("Bot " + os.Getenv("DISCORD_KEY"))
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
		saveGuildInfo(discord, message.GuildID, db.Connection)
		discord.ChannelMessageSendReply(message.ChannelID, "Done", message.Reference())

	case strings.HasPrefix(message.Content, "%%danceInEveryChannel"):
		dance(discord, message, "")

	case strings.HasPrefix(message.Content, "%%danceHere"):
		dance(discord, message, message.ChannelID)

	case strings.HasPrefix(message.Content, "%%helpMeRankEmojis"):
		discord.ChannelMessageSend(message.ChannelID, "This is an example, figure it out: %%rankUsedEmojisInGuild author=123 channel=123321 ignoreReactions=true belongToTheGuild=false ignoreMessageText=false fromDate=2022-01-01 toDate=2024-01-01 desc=true limit=10")
	case strings.HasPrefix(message.Content, "%%helpMeRankReactions"):
		discord.ChannelMessageSend(message.ChannelID, "This is a special case messageAuthor only works like this(dates are optional): %%rankUsedEmojisInGuild messageAuthor=123 ignoreMessageText=true fromDate=2022-01-01 toDate=2024-01-01 desc=true limit=10")
	case strings.HasPrefix(message.Content, rankUsedEmojisInGuild):
		rankingHandler(discord, message)
	}

	err := ProcessOneMessage(nil, MessageModel{Message: message.Message}, message.GuildID, db.Connection, true)

	NotifyAboutErrorViaWebhook(err)

	return
}
