package easter

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"math/rand"
	"os"
	"regexp"
	"strings"
)

func ProcessEasterEgg(discord *discordgo.Session, message *discordgo.MessageCreate) {
	if message.Author.Bot {
		return
	}

	processEasterEgg2(discord, message)

	if message.Author.ID != os.Getenv("EASTER_EGG_AUTHOR_ID") {
		return
	}

	rx := regexp.MustCompile(os.Getenv("EASTER_EGG_REGEX"))

	if rx.MatchString(strings.Trim(message.Content, " ")) {
		options := strings.Split(os.Getenv("EASTER_EGG_LINK_OPTIONS"), ",")

		discord.ChannelMessageSendReply(message.ChannelID, options[rand.Intn(len(options))], message.Reference())
	}
}

func processEasterEgg2(discord *discordgo.Session, message *discordgo.MessageCreate) {
	if len(message.Mentions) == 1 && message.Mentions[0].ID == os.Getenv("EASTER_EGG2_USER_ID") {
		if strings.Trim(message.Message.Content, " ") == fmt.Sprintf("<@%s>", os.Getenv("EASTER_EGG2_USER_ID")) {
			discord.ChannelMessageSend(message.ChannelID, os.Getenv("EASTER_EGG2_CONTENT"))
		}
	}
}
