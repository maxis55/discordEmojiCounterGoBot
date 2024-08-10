package bot

import (
	"database/sql"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"slices"
)

const messagesLimit = 100

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

		if len(message.Reactions) > 0 {
			var reactModels []EmojiModel
			for _, reaction := range message.Reactions {
				users, err := discord.MessageReactions(channel.ID, message.ID, reaction.Emoji.APIName(), 100, "", "", requestConfig)
				if err != nil {
					fmt.Println(err.Error())
				}

				for _, user := range users {
					if _, ok := authors[user.ID]; !ok {
						authors[user.ID] = AuthorModel{Author: user}
					}
				}

				reactModels = slices.Concat(reactModels, getReactionsAsModels(users, reaction.Emoji))
			}
			mm.Reactions = reactModels
		}

		ProcessOneMessage(mm, gid, db, false)
	}

	if len(authors) > 10 {
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
