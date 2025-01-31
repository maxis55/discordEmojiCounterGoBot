package bot

import (
	"database/sql"
	"discordEmojiCounterBot/utils"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strings"
)

type RankedEmojis []RankedEmoji

type RankedEmoji struct {
	Count int
	Rank  int
	Emoji discordgo.Emoji
}

const maxEmbedFieldValueLen = 1024
const maxEmbedTitleLen = 256

func getRankedUsedEmojisInGuild(db *sql.DB, gid string, settings RankingSettings) (RankedEmojis, error) {
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

	if settings.MessageAuthorId != nil && *settings.MessageAuthorId != "" {
		paramsC++
		query += fmt.Sprintf(" and eu.m_author_id = $%d", paramsC)
		params = append(params, *settings.MessageAuthorId)
	}

	if settings.IgnoreReactions != nil && *settings.IgnoreReactions {
		query += " and eu.is_reaction != true"
	}

	if settings.IgnoreMessageText != nil && *settings.IgnoreMessageText {
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

	order := "desc"

	if settings.Desc != nil && !*settings.Desc {
		order = "asc"
	}

	query += fmt.Sprintf(" order by emoji_count %s", order)

	if settings.Limit != nil && *settings.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", *settings.Limit)
	}

	rows, err := db.Query(query, params...)

	if err != nil {
		return nil, err
	}

	currRank := 1
	var rankedEmojis RankedEmojis

	for rows.Next() {
		rankedEmoji := RankedEmoji{Emoji: discordgo.Emoji{}, Rank: currRank}
		err = rows.Scan(&rankedEmoji.Emoji.ID, &rankedEmoji.Emoji.Name, &rankedEmoji.Emoji.Animated, &rankedEmoji.Count)

		if emojiRegex.MatchString(rankedEmoji.Emoji.Name) {
			rankedEmoji.Emoji.ID = ""
		}

		if err != nil {
			return nil, err
		}

		rankedEmojis = append(rankedEmojis, rankedEmoji)
		currRank++
	}

	return rankedEmojis, nil
}

//max embed size is 6000
//each embed field has max value len of 1024
//by sending max 1024*5, there's room to spare for settings(in title, 256 symbols), etc
//send *columns* fields in each embed message, this way there's always room to spare

func (res *RankedEmojis) TransformIntoDiscordEmbeds(settingsJS string) []discordgo.MessageEmbed {
	if len(*res) == 0 {
		return []discordgo.MessageEmbed{{Title: settingsJS, Description: "No results."}}
	}

	var embeds []discordgo.MessageEmbed

	const columns = 5

	formattedSlice := res.getFormattedSlice()

	slicedRankedEmojis := utils.ChunkSliceValuesByLen(formattedSlice, maxEmbedFieldValueLen)

	embedCount := 0
	var embed discordgo.MessageEmbed

	for columnIndex, column := range slicedRankedEmojis {
		if columnIndex%columns == 0 {
			if len(embed.Fields) > 0 {
				embeds = append(embeds, embed)
			}

			title := settingsJS
			if len(slicedRankedEmojis) > columns {
				title = fmt.Sprintf("%s #%d", title, embedCount+1)
			}

			embed = discordgo.MessageEmbed{Title: title}
			embedCount++
		}

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Value: strings.Join(column, ""), Inline: true})
	}
	embeds = append(embeds, embed)

	return embeds

}

func (re *RankedEmoji) getEmbedContent() string {
	return fmt.Sprintf("%d. %s - %d", re.Rank, re.Emoji.MessageFormat(), re.Count)
}

func (res *RankedEmojis) getFormattedSlice() []string {
	var result []string

	for _, re := range *res {
		result = append(result, re.getEmbedContent()+"\n")
	}

	return result
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
