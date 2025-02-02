package emoji_processing

import (
	"database/sql"
	"github.com/bwmarrin/discordgo"
)

func ProcessOneMessage(discord *discordgo.Session, message MessageModel, gid string, db *sql.DB) error {
	if message.Message.Author.Bot {
		return nil
	}

	ejs, err := message.GetEmojisFromMessage(discord)

	if err != nil {
		return err
	}

	//js, _ := json.Marshal(m)
	//fmt.Println(string(js))

	//save to DB
	rememberNewEmojis(ejs, db)

	err = cleanEmojiInfoAboutMessage(message.Message.ID, db)
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

	err = AuthorModel{Author: message.Message.Author}.remember(db)
	if err != nil {
		return err
	}

	return nil
}

func CleanInfoAboutMessage(mid string, db *sql.DB) error {
	if _, err := db.Exec("DELETE FROM messages WHERE message_id=$1;", mid); err != nil {
		return err
	}

	if err := cleanEmojiInfoAboutMessage(mid, db); err != nil {
		return err
	}

	return nil
}

func cleanEmojiInfoAboutMessage(mid string, db *sql.DB) error {
	if _, err := db.Exec("DELETE FROM emoji_used where message_id=$1;", mid); err != nil {
		return err
	}
	return nil
}
