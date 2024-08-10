package bot

import (
	"database/sql"
	"fmt"
)

func ProcessOneMessage(message MessageModel, gid string, db *sql.DB, saveAuthor bool) {
	if message.Message.Author.Bot {
		return
	}

	ejs := getEmojisFromMessage(message)

	//js, _ := json.Marshal(m)
	//fmt.Println(string(js))

	//save to DB
	saveEmojis(ejs, db)

	//createEmojiCounts
	err := message.remember(gid, db)
	if err != nil {
		fmt.Println(err.Error())
	}

	err = message.saveEmojiUsages(db, ejs, gid)
	if err != nil {
		fmt.Println(err.Error())
	}

	if saveAuthor {
		err = AuthorModel{Author: message.Message.Author}.remember(db)
		if err != nil {
			fmt.Println(err.Error())
		}
	}

}
