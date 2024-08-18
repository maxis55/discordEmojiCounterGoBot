package bot

import (
	"database/sql"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

const messagesLimit = 100

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

func saveGuildInfo(discord *discordgo.Session, gid string, db *sql.DB) {
	guild, err := discord.Guild(gid)
	if err != nil {
		fmt.Println(err.Error())
	}

	err = GuildModel{Guild: guild}.remember(db)

	if err != nil {
		fmt.Println(err.Error())
	}

	ejModels := make([]EmojiModel, 0, len(guild.Emojis))

	for _, emoji := range guild.Emojis {
		if emoji == nil {
			continue
		}
		ejModels = append(ejModels, EmojiModel{
			Emoji:   *emoji,
			GuildID: &gid,
		})

	}

	saveEmojis(ejModels, db)

	channels, err := discord.GuildChannels(gid)
	if err != nil {
		fmt.Println(err.Error())
	}

	for _, channel := range channels {
		err = ChannelModel{Channel: channel}.remember(db)
		if err != nil {
			fmt.Println(err.Error())
		}
	}

}

func requestConfig(cfg *discordgo.RequestConfig) {
	cfg.ShouldRetryOnRateLimit = true
	cfg.MaxRestRetries = 15
}

func getAndSaveAllMessages(discord *discordgo.Session, bMessage *discordgo.Message, gid string, db *sql.DB, authors map[string]AuthorModel, reference *discordgo.MessageReference, channel *discordgo.Channel) {
	ms, err := discord.ChannelMessages(channel.ID, messagesLimit, bMessage.ID, "", "", func(cfg *discordgo.RequestConfig) {
		cfg.ShouldRetryOnRateLimit = true
		cfg.MaxRestRetries = 15
	})

	if err != nil {
		errMsg := err.Error()
		_, err = discord.ChannelMessageSendReply(reference.ChannelID, "💀 Reason: "+errMsg+" channel "+channel.Name, reference, requestConfig)

		fmt.Println(errMsg)
		return
	}

	for _, message := range ms {
		if _, ok := authors[message.Author.ID]; !ok {
			authors[message.Author.ID] = AuthorModel{Author: message.Author}
		}

		mm := MessageModel{Message: message}

		err = ProcessOneMessage(discord, mm, gid, db, false)
		if err != nil {
			errMsg := err.Error()
			_, err = discord.ChannelMessageSendReply(reference.ChannelID, "💀 Reason: "+errMsg+" channel "+channel.Name, reference, requestConfig)

			fmt.Println(errMsg)
			return
		}
	}

	if len(authors) > 20 {
		saveAuthors(authors, db)
		clear(authors)
	}

	if len(ms) < messagesLimit {
		saveAuthors(authors, db)

		_, err = discord.ChannelMessageSendReply(reference.ChannelID, fmt.Sprintf("Finished parsing channel '%s'", channel.Name), reference, requestConfig)

		if err != nil {
			fmt.Println(err.Error())
		}
		return
	}

	getAndSaveAllMessages(discord, ms[len(ms)-1], gid, db, authors, reference, channel)

}

func saveAuthors(authors map[string]AuthorModel, db *sql.DB) {
	for _, author := range authors {
		err := author.remember(db)
		if err != nil {
			fmt.Println(err.Error())
		}
	}
}

func danceInEveryChannel(discord *discordgo.Session, message *discordgo.MessageCreate) {
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
}

func danceHere(discord *discordgo.Session, message *discordgo.MessageCreate) {
	trackingMsg, err := discord.ChannelMessageSend(message.ChannelID, "Ok"+dancers[rand.Intn(len(dancers))])

	if err != nil {
		discord.ChannelMessageSendReply(message.ChannelID, "💀 Reason: "+err.Error(), message.Reference(), requestConfig)
		return
	}

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
}
