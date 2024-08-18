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

func getRankedUsedEmojisInGuild(db *sql.DB, gid string, settings RankingSettings) (string, error) {

	params := []interface{}{
		gid,
	}
	paramsC := 1

	query := `select e.emoji_id, e.name, e.animated, COUNT(eu.id) as emoji_count
									from emoji_used eu join emojis e on eu.emoji_id = e.emoji_id
									where eu.guild_id = $1`

	if settings.BelongToTheGuild != nil && *settings.BelongToTheGuild {
		query += " and e.guild_id = $1"
	}

	if settings.ChannelID != nil && *settings.ChannelID != "" {
		paramsC++
		query += fmt.Sprintf(" and eu.channel_id = $%d", paramsC)
		params = append(params, *settings.ChannelID)

	}

	if settings.AuthorID != nil && *settings.AuthorID != "" {
		paramsC++
		query += fmt.Sprintf(" and eu.author_id = $%d", paramsC)
		params = append(params, *settings.AuthorID)
	}

	if settings.WithoutReactions != nil && *settings.WithoutReactions {
		query += " and eu.is_reaction != true"
	}

	if settings.WithoutMessageText != nil && *settings.WithoutMessageText {
		query += " and eu.is_reaction != false"
	}

	if settings.WithoutMessageText != nil && *settings.WithoutMessageText {
		query += " and eu.is_reaction != false"
	}

	if settings.FromDate != nil && *settings.FromDate != "" {
		paramsC++
		query += fmt.Sprintf(" and eu.timestamp >= $%d::timestamp", paramsC)
		params = append(params, *settings.FromDate)
	}

	if settings.ToDate != nil && *settings.ToDate != "" {
		paramsC++
		query += fmt.Sprintf(" and eu.timestamp <= $%d::timestamp", paramsC)
		params = append(params, *settings.ToDate)
	}

	query += " group by e.emoji_id"

	if settings.Desc != nil {
		order := "desc"

		if !*settings.Desc {
			order = "asc"
		}

		query += fmt.Sprintf(" order by emoji_count %s", order)
	}

	if settings.Limit != nil && *settings.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", *settings.Limit)
	}

	rows, err := db.Query(query, params...)
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

	if result == "" {
		result = "Empty set."
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
