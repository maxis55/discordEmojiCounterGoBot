package emoji_processing

import (
	"database/sql"
	"emoji-counter/utils"
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

func GetRankedUsedEmojisInGuild(db *sql.DB, gid string, rs RankingSettings) (RankedEmojis, error) {
	var params []interface{}
	var queryBuilder strings.Builder

	params = append(params, gid)

	queryBuilder.WriteString(`select e.emoji_id, e.name, e.animated, COUNT(eu.id) as emoji_count
		from emoji_used eu join emojis e on eu.emoji_id = e.emoji_id
		where eu.guild_id = $1`)

	var conditions []string

	if rs.BelongToTheGuild != nil && *rs.BelongToTheGuild {
		conditions = append(conditions, "e.guild_id = $1")
	}

	if rs.ChannelID != nil && *rs.ChannelID != "" {
		params = append(params, *rs.ChannelID)
		conditions = append(conditions, fmt.Sprintf("eu.channel_id = $%d", len(params)))
	}

	if rs.AuthorID != nil && *rs.AuthorID != "" {
		params = append(params, *rs.AuthorID)
		conditions = append(conditions, fmt.Sprintf("eu.author_id = $%d", len(params)))
	}

	if rs.MessageAuthorId != nil && *rs.MessageAuthorId != "" {
		params = append(params, *rs.MessageAuthorId)
		conditions = append(conditions, fmt.Sprintf("eu.m_author_id = $%d", len(params)))
	}

	if rs.IgnoreReactions != nil && *rs.IgnoreReactions {
		conditions = append(conditions, "eu.is_reaction != true")
	}

	if rs.IgnoreMessageText != nil && *rs.IgnoreMessageText {
		conditions = append(conditions, "eu.is_reaction != false")
	}

	if rs.FromDate != nil && *rs.FromDate != "" {
		params = append(params, *rs.FromDate)
		conditions = append(conditions, fmt.Sprintf("eu.timestamp >= $%d::timestamp", len(params)))
	}

	if rs.ToDate != nil && *rs.ToDate != "" {
		params = append(params, *rs.ToDate)
		conditions = append(conditions, fmt.Sprintf("eu.timestamp <= $%d::timestamp", len(params)))
	}

	if len(conditions) > 0 {
		queryBuilder.WriteString(" and " + strings.Join(conditions, " and "))
	}

	queryBuilder.WriteString(" group by e.emoji_id")

	order := "desc"
	if rs.Desc != nil && !*rs.Desc {
		order = "asc"
	}

	queryBuilder.WriteString(fmt.Sprintf(" order by emoji_count %s", order))

	if rs.Limit != nil && *rs.Limit > 0 {
		queryBuilder.WriteString(fmt.Sprintf(" LIMIT %d", *rs.Limit))
	}

	rows, err := db.Query(queryBuilder.String(), params...)

	if err != nil {
		return nil, err
	}

	currRank := 1
	var rankedEmojis RankedEmojis

	for rows.Next() {
		rankedEmoji := RankedEmoji{Emoji: discordgo.Emoji{}, Rank: currRank}
		err = rows.Scan(&rankedEmoji.Emoji.ID, &rankedEmoji.Emoji.Name, &rankedEmoji.Emoji.Animated, &rankedEmoji.Count)

		if err != nil {
			return nil, err
		}

		if emojiRegex.MatchString(rankedEmoji.Emoji.Name) {
			rankedEmoji.Emoji.ID = ""
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
//before sending - redistribute between 6 fields to make messages more event
//max emoji len is 32 symbols, so there's no edge cases with some slices in the middle being overwhelmed on redistribution

func (res *RankedEmojis) TransformIntoDiscordEmbeds(settingsJS string) []discordgo.MessageEmbed {
	if len(*res) == 0 {
		return []discordgo.MessageEmbed{{Title: settingsJS, Description: "No results."}}
	}

	var embeds []discordgo.MessageEmbed
	embedCount := 0
	var embed discordgo.MessageEmbed

	const maxColumnsBasedOnChars = 5
	const redistributedColumnsCount = 6
	formattedSlice := res.getFormattedSlice()
	slicedRankedEmojis := utils.ChunkSliceValuesByLen(formattedSlice, maxEmbedFieldValueLen)
	chunkedRankedEmojis := utils.ChunkSlice(slicedRankedEmojis, maxColumnsBasedOnChars)

	for _, chunk := range chunkedRankedEmojis {
		redistributedSlices := utils.RedistributeSlicesIntoAmountBasedOnEntries(chunk, redistributedColumnsCount)
		redistributedSlices = utils.RedistributeSlicesBasedOnMaxLen(redistributedSlices, maxEmbedFieldValueLen)

		for columnIndex, column := range redistributedSlices {
			if columnIndex%redistributedColumnsCount == 0 {
				if len(embed.Fields) > 0 {
					embeds = append(embeds, embed)
				}

				embedCount++
				title := settingsJS

				if len(slicedRankedEmojis) > maxColumnsBasedOnChars {
					title = fmt.Sprintf("%s #%d", title, embedCount)
				}

				embed = discordgo.MessageEmbed{Title: title}
			}

			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Value: strings.Join(column, ""), Inline: true})
		}

	}

	embeds = append(embeds, embed)

	return embeds
}

func (re *RankedEmoji) getEmbedContent() string {
	return fmt.Sprintf("**%d**. %s - %d", re.Rank, re.Emoji.MessageFormat(), re.Count)
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
