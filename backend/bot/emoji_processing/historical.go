package emoji_processing

import (
	"database/sql"
	"emoji-counter/db"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log/slog"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

const messagesPerPage = 100

var dancers = [...]string{"💃", "💃🏻", "💃🏼", "💃🏽", "💃🏾", "💃🏿", "🕺🏿", "🕺🏾", "🕺🏽", "🕺🏼", "🕺🏻", "🕺"}

type waitGroupCount struct {
	sync.WaitGroup
	count int64
}

func (wg *waitGroupCount) Add(delta int) {
	atomic.AddInt64(&wg.count, int64(delta))
	wg.WaitGroup.Add(delta)
}

func (wg *waitGroupCount) Done() {
	atomic.AddInt64(&wg.count, -1)
	wg.WaitGroup.Done()
}

func (wg *waitGroupCount) GetCount() int {
	return int(atomic.LoadInt64(&wg.count))
}

func RequestConfig(cfg *discordgo.RequestConfig) {
	cfg.ShouldRetryOnRateLimit = true
	cfg.MaxRestRetries = 15
}

func getAndSaveAllMessages(discord *discordgo.Session, beforeMessage *discordgo.Message, gid string, db *sql.DB, waiter *discordgo.MessageReference, c *discordgo.Channel) {
	ms, err := discord.ChannelMessages(c.ID, messagesPerPage, beforeMessage.ID, "", "", RequestConfig)

	if err != nil {
		slog.Error("fetch channel messages failed", "channel", c.Name, "err", err)
		_, _ = discord.ChannelMessageSendReply(waiter.ChannelID, "💀 Reason: "+err.Error()+" channel "+c.Name, waiter, RequestConfig)
		return
	}

	for _, message := range ms {
		mm := MessageModel{Message: message}

		err = ProcessOneMessage(discord, mm, gid, db)
		if err != nil {
			slog.Error("process historical message failed", "channel", c.Name, "err", err)
			_, _ = discord.ChannelMessageSendReply(waiter.ChannelID, "💀 Reason: "+err.Error()+" channel "+c.Name, waiter, RequestConfig)
			return
		}
	}

	if len(ms) < messagesPerPage {
		_, err = discord.ChannelMessageSendReply(waiter.ChannelID, fmt.Sprintf("Finished parsing channel '%s'", c.Name), waiter, RequestConfig)

		if err != nil {
			NotifyAboutErrorViaWebhook(err)
		}
		return
	}

	getAndSaveAllMessages(discord, ms[len(ms)-1], gid, db, waiter, c)
}

func Dance(discord *discordgo.Session, message *discordgo.MessageCreate, channels []ChannelModel) {
	trackingMsg, _ := discord.ChannelMessageSend(message.ChannelID, "Ok"+dancers[rand.Intn(len(dancers))])

	var err error
	wg := waitGroupCount{}

	for _, channel := range channels {
		wg.Add(1)
		go func(ch ChannelModel) {
			defer wg.Done()
			getAndSaveAllMessages(discord, message.Message, message.GuildID, db.Connection, message.Reference(), ch.Channel)
		}(channel)
	}
	for wg.GetCount() > 0 {
		trackingMsg, err = discord.ChannelMessageEdit(trackingMsg.ChannelID, trackingMsg.ID, fmt.Sprintf("Working on %d channels", wg.GetCount()), RequestConfig)
		NotifyAboutErrorViaWebhook(err)
		time.Sleep(time.Second * 2)
	}
	discord.ChannelMessageEdit(trackingMsg.ChannelID, trackingMsg.ID, "Done", RequestConfig)
}
