package easter

import (
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

	if message.Author.ID != os.Getenv("EASTER_EGG_AUTHOR_ID") {
		return
	}

	rx := regexp.MustCompile(os.Getenv("EASTER_EGG_REGEX"))

	if rx.MatchString(strings.Trim(message.Content, " ")) {
		options := strings.Split(os.Getenv("EASTER_EGG_LINK_OPTIONS"), ",")

		discord.ChannelMessageSendReply(message.ChannelID, options[rand.Intn(len(options))], message.Reference())
	}
}
