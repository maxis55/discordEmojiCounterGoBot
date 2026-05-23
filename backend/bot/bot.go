package bot

import (
	"context"
	"emoji-counter/bot/easter"
	"emoji-counter/bot/emoji_processing"
	"emoji-counter/db"
	"log/slog"
	"os"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
)

const pr = "%%"
const rankUsedEmojisInGuild = pr + "rankUsedEmojisInGuild"

var wg sync.WaitGroup

// goAsync runs fn in a goroutine tracked by the package WaitGroup so that
// Run() can drain in-flight handler work before shutdown.
func goAsync(fn func()) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		fn()
	}()
}

// guildSaveInflight tracks guild IDs that have an in-progress SaveGuildInfo
// goroutine, so a double-trigger of %%saveEverythingAboutThisGuild doesn't
// kick off two simultaneous imports for the same guild.
var guildSaveInflight sync.Map

func Run(ctx context.Context) {
	discord, err := discordgo.New("Bot " + os.Getenv("DISCORD_KEY"))
	if err != nil {
		slog.Error("discord session create failed", "err", err)
		os.Exit(1)
	}

	discord.AddHandler(newMessage)
	discord.AddHandler(messageUpdated)
	discord.AddHandler(messageDeleted)
	discord.AddHandler(newReaction)
	discord.AddHandler(removedReaction)
	discord.AddHandler(removedAllReactions)
	discord.AddHandler(rateLimited)

	if err := discord.Open(); err != nil {
		emoji_processing.NotifyAboutErrorViaWebhook(err)
		slog.Error("discord session open failed", "err", err)
		os.Exit(1)
	}
	defer discord.Close()

	slog.Info("bot running")
	<-ctx.Done()
	slog.Info("shutdown signal received, draining handlers")
	wg.Wait()
	slog.Info("handlers drained")
}

func newMessage(discord *discordgo.Session, message *discordgo.MessageCreate) {
	if message.Author.ID == discord.State.User.ID || message.Author.Bot {
		return
	}

	switch {
	case strings.HasPrefix(message.Content, "%%hello"):
		goAsync(func() { discord.ChannelMessageSend(message.ChannelID, "Hello World😃") })
	case strings.HasPrefix(message.Content, "%%bye"):
		goAsync(func() { discord.ChannelMessageSend(message.ChannelID, "Good Bye👋") })
	case strings.HasPrefix(message.Content, "%%saveEverythingAboutThisGuild"):
		if _, loaded := guildSaveInflight.LoadOrStore(message.GuildID, struct{}{}); loaded {
			goAsync(func() {
				discord.ChannelMessageSend(message.ChannelID, "Already saving this guild, please wait")
			})
			break
		}
		goAsync(func() {
			defer guildSaveInflight.Delete(message.GuildID)
			discord.ChannelMessageSend(message.ChannelID, "Ok")
			emoji_processing.SaveGuildInfo(discord, message.GuildID, db.Connection)
			discord.ChannelMessageSendReply(message.ChannelID, "Done", message.Reference())
		})
	case strings.HasPrefix(message.Content, "%%danceInEveryChannel"):
		goAsync(func() { handleHistoricalForGuild(discord, message) })
	case strings.HasPrefix(message.Content, "%%danceHere"):
		goAsync(func() { handleHistoricalForChannel(discord, message) })
	case strings.HasPrefix(message.Content, "%%helpMeRankEmojis"):
		goAsync(func() {
			discord.ChannelMessageSend(message.ChannelID, "This is an example, figure it out: %%rankUsedEmojisInGuild author=123 channel=123321 ignoreReactions=true belongToTheGuild=false ignoreMessageText=false fromDate=2022-01-01 toDate=2024-01-01 desc=true limit=10")
		})
	case strings.HasPrefix(message.Content, "%%helpMeRankReactions"):
		goAsync(func() {
			discord.ChannelMessageSend(message.ChannelID, "This is a special case messageAuthor only works like this(dates are optional): %%rankUsedEmojisInGuild messageAuthor=123 ignoreMessageText=true fromDate=2022-01-01 toDate=2024-01-01 desc=true limit=10")
		})
	case strings.HasPrefix(message.Content, rankUsedEmojisInGuild):
		goAsync(func() { rankingHandler(discord, message) })
	}

	goAsync(func() {
		err := emoji_processing.ProcessOneMessage(discord, emoji_processing.MessageModel{Message: message.Message}, message.GuildID, db.Connection)
		emoji_processing.NotifyAboutErrorViaWebhook(err)
	})

	goAsync(func() {
		easter.ProcessEasterEgg(discord, message)
	})
}
