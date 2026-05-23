package emoji_processing

import (
	"database/sql"
	dbpkg "emoji-counter/db"
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

	// Save parents before children so referential intent holds (no FKs today, but
	// keeps the door open and makes the data shape less surprising).
	if err := (AuthorModel{Author: message.Message.Author}).Remember(db); err != nil {
		return err
	}

	rememberNewEmojis(ejs, db)

	if err := forgetUsedEmojisRelatedToMessage(message.Message.ID, db); err != nil {
		return err
	}

	if err := message.remember(gid, db); err != nil {
		return err
	}

	if err := message.saveEmojiUsages(db, ejs, gid); err != nil {
		return err
	}

	return nil
}

func ForgetEverythingAboutMessage(mid string, db *sql.DB) error {
	if _, err := dbpkg.Exec(db, "DELETE FROM messages WHERE message_id=$1;", mid); err != nil {
		return err
	}

	if err := forgetUsedEmojisRelatedToMessage(mid, db); err != nil {
		return err
	}

	return nil
}

func forgetUsedEmojisRelatedToMessage(mid string, db *sql.DB) error {
	if _, err := dbpkg.Exec(db, "DELETE FROM emoji_used where message_id=$1;", mid); err != nil {
		return err
	}
	return nil
}
