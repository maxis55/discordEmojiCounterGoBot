package bot

import (
	"database/sql"
	"github.com/bwmarrin/discordgo"
	"slices"
)

func ProcessOneMessage(discord *discordgo.Session, message MessageModel, gid string, db *sql.DB, saveRightAway bool) error {
	if message.Message.Author.Bot {
		return nil
	}

	reacts, err := populateReactionEmojis(discord, message)

	if err != nil {
		return err
	}

	message.Reactions = reacts

	ejs := getEmojisFromMessage(message)

	//js, _ := json.Marshal(m)
	//fmt.Println(string(js))

	//save to DB
	saveEmojis(ejs, db)

	err = cleanInfoAboutMessage(message.Message.ID, db)
	if err != nil {
		return err
	}

	err = message.remember(gid, db)
	if err != nil {
		return err
	}

	err = message.saveEmojiUsages(db, ejs, gid)
	if err != nil {
		return err
	}

	if saveRightAway {
		err = AuthorModel{Author: message.Message.Author}.remember(db)
		if err != nil {
			return err
		}
	}

	return nil
}

func populateReactionEmojis(discord *discordgo.Session, message MessageModel) ([]EmojiModel, error) {
	if len(message.Message.Reactions) > 0 {
		var reactModels []EmojiModel
		for _, reaction := range message.Message.Reactions {
			users, err := discord.MessageReactions(message.Message.ChannelID, message.Message.ID, reaction.Emoji.APIName(), 100, "", "", requestConfig)
			if err != nil {
				return nil, err
			}

			reactModels = slices.Concat(reactModels, getReactionsAsModels(users, reaction.Emoji))
		}
		return reactModels, nil
	}
	return nil, nil
}
