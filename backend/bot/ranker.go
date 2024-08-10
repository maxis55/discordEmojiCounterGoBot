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

func getRankedUsedEmojisInGuild(db *sql.DB, gid string, limit int, desc bool) (string, error) {

	query := `select e.emoji_id, e.name, e.animated, COUNT(eu.id) as emoji_count
									from emoji_used eu join emojis e on eu.emoji_id = e.emoji_id
									where eu.guild_id = $1
									group by e.emoji_id`

	order := "asc"

	if desc {
		order = " asc"
	}

	query += fmt.Sprintf(" order by emoji_count %s", order)
	query += fmt.Sprintf(" LIMIT %d", limit)

	rows, err := db.Query(query, gid)
	if err != nil {
		return "", err
	}

	var models []EmojiRank

	for rows.Next() {
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

func getRankedAvailableUsedEmojisInGuild(db *sql.DB, gid string, limit int, desc bool) (string, error) {

	query := `select e.emoji_id, e.name, e.animated, COUNT(eu.id) as emoji_count
									from emoji_used eu join emojis e on eu.emoji_id = e.emoji_id
									where eu.guild_id = $1 and e.guild_id = $1
									group by e.emoji_id`
	order := "asc"

	if desc {
		order = " asc"
	}

	query += fmt.Sprintf(" order by emoji_count %s", order)
	query += fmt.Sprintf(" LIMIT %d", limit)

	rows, err := db.Query(query, gid)

	if err != nil {
		return "", err
	}

	var models []EmojiRank

	for rows.Next() {
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

//select a.username, count(*) as blabbing_messages
//from messages m
//join public.authors a on m.author_id = a.author_id
//where a.author_id in ()
//and m.timestamp>='2022-01-01 01:00:00.307000'
//group by a.author_id;
//
//
//select a.username, SUM(CHAR_LENGTH(m.content)) as blabbing_symbols
//from messages m
//join public.authors a on m.author_id = a.author_id
//where a.author_id in ()
//and m.timestamp>='2022-01-01 01:00:00.307000'
//group by a.author_id
