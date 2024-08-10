package bot

import (
	"database/sql"
)

func ProcessOneMessage(message MessageModel, gid string, db *sql.DB, saveAuthor bool) error {
	if message.Message.Author.Bot {
		return nil
	}

	ejs := getEmojisFromMessage(message)

	//js, _ := json.Marshal(m)
	//fmt.Println(string(js))

	//save to DB
	saveEmojis(ejs, db)

	err := cleanInfoAboutMessage(message.Message.ID, db)
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

	if saveAuthor {
		err = AuthorModel{Author: message.Message.Author}.remember(db)
		if err != nil {
			return err
		}
	}

	return nil
}
