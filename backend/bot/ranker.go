package bot

import (
	"database/sql"
	"fmt"
	"github.com/bwmarrin/discordgo"
)

type EmojiRank struct {
	Count int
	Emoji discordgo.Emoji
}

func rankUsedEmojisInGuild(db *sql.DB, gid string, limit int) (string, error) {
	rows, err := db.Query(`select e.emoji_id, e.name, e.animated, COUNT(eu.id) as emoji_count
									from emoji_used eu
											 join emojis e on eu.emoji_id = e.emoji_id
									where eu.guild_id = $1
									group by e.emoji_id
									order by emoji_count desc`,
		gid)
	if err != nil {
		return "", err
	}

	var models []EmojiRank

	for rows.Next() {

		if limit > 0 && len(models) >= limit {
			break
		}

		model := EmojiRank{Emoji: discordgo.Emoji{}}
		err = rows.Scan(&model.Emoji.ID, &model.Emoji.Name, &model.Emoji.Animated, &model.Count)

		if emojiRegex.MatchString(model.Emoji.Name) {
			model.Emoji.ID = ""
		}

		if err != nil {
			return "", err
		}

		models = append(models, model)
	}

	var result string

	for i, model := range models {
		result += fmt.Sprintf("%d. %s - %d \n", i+1, model.Emoji.MessageFormat(), model.Count)
	}

	return result, nil
}
