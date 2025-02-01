package bot

import (
	"database/sql"
	"emoji-counter/db"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

const messagesPerPage = 100

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

func requestConfig(cfg *discordgo.RequestConfig) {
	cfg.ShouldRetryOnRateLimit = true
	cfg.MaxRestRetries = 15
}

func getAndSaveAllMessages(discord *discordgo.Session, beforeMessage *discordgo.Message, gid string, db *sql.DB, waiter *discordgo.MessageReference, c *discordgo.Channel) {
	ms, err := discord.ChannelMessages(c.ID, messagesPerPage, beforeMessage.ID, "", "", requestConfig)

	if err != nil {
		errMsg := err.Error()
		_, err = discord.ChannelMessageSendReply(waiter.ChannelID, "💀 Reason: "+errMsg+" channel "+c.Name, waiter, requestConfig)

		fmt.Println(errMsg)
		return
	}

	for _, message := range ms {
		mm := MessageModel{Message: message}

		err = ProcessOneMessage(discord, mm, gid, db)
		if err != nil {
			errMsg := err.Error()
			_, err = discord.ChannelMessageSendReply(waiter.ChannelID, "💀 Reason: "+errMsg+" channel "+c.Name, waiter, requestConfig)

			fmt.Println(errMsg)
			return
		}
	}

	if len(ms) < messagesPerPage {
		_, err = discord.ChannelMessageSendReply(waiter.ChannelID, fmt.Sprintf("Finished parsing channel '%s'", c.Name), waiter, requestConfig)

		if err != nil {
			NotifyAboutErrorViaWebhook(err)
		}
		return
	}

	getAndSaveAllMessages(discord, ms[len(ms)-1], gid, db, waiter, c)
}

func dance(discord *discordgo.Session, message *discordgo.MessageCreate, channelID string) {
	trackingMsg, _ := discord.ChannelMessageSend(message.ChannelID, "Ok"+dancers[rand.Intn(len(dancers))])

	var channels []ChannelModel
	var err error

	if channelID != "" {
		channel, err := queryChannelById(db.Connection, message.ChannelID)

		if err != nil {
			discord.ChannelMessageSendReply(message.ChannelID, "💀 Reason: "+err.Error(), message.Reference(), requestConfig)
			return
		}

		if channel == nil {
			discord.ChannelMessageSendReply(message.ChannelID, "Cant find the channel in the DB. Save this guild first maybe", message.Reference(), requestConfig)
			return
		}

		channels = append(channels, *channel)
	}

	if len(channels) < 1 {
		channels, err = queryAllGuildChannels(db.Connection, message.GuildID)
		if err != nil {
			discord.ChannelMessageSendReply(message.ChannelID, "💀 Reason: "+err.Error(), message.Reference(), requestConfig)
			return
		}
	}

	wg := WaitGroupCount{}

	for _, channel := range channels {
		wg.Add(1)
		go func(ch ChannelModel) {
			defer wg.Done()
			getAndSaveAllMessages(discord, message.Message, message.GuildID, db.Connection, message.Reference(), ch.Channel)
		}(channel)
	}
	for wg.GetCount() > 0 {
		trackingMsg, err = discord.ChannelMessageEdit(trackingMsg.ChannelID, trackingMsg.ID, fmt.Sprintf("Working on %d channels", wg.GetCount()), requestConfig)
		time.Sleep(time.Second * 2)
	}
	discord.ChannelMessageEdit(trackingMsg.ChannelID, trackingMsg.ID, "Done", requestConfig)
}
