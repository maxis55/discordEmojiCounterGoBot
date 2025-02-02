package emoji_processing

import (
	"database/sql"
	"emoji-counter/utils"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strings"
)

type MessageModel struct {
	Message *discordgo.Message
}

func (mm *MessageModel) remember(gid string, db *sql.DB) error {
	content := mm.Message.Content

	if utils.GetEnvBoolWithFallback("REMEMBER_MESSAGE_CONTENT", false) {
		content = ""
	}

	valuesMap := map[string]any{
		"message_id":       mm.Message.ID,
		"content":          content,
		"guild_id":         gid,
		"channel_id":       mm.Message.ChannelID,
		"type":             mm.Message.Type,
		"author_id":        mm.Message.Author.ID,
		"timestamp":        mm.Message.Timestamp,
		"edited_timestamp": mm.Message.EditedTimestamp,
	}

	fields := utils.GetKeysFromMap(valuesMap)
	values := utils.GetValuesFromMapBasedOnKeys(valuesMap, fields)

	query := fmt.Sprintf(`
		INSERT INTO messages (%s)
		VALUES (%s)
		ON CONFLICT (message_id) DO UPDATE
		SET content=EXCLUDED.content, edited_timestamp=EXCLUDED.edited_timestamp;
	`, strings.Join(fields, ", "), utils.SQLPlaceholders(len(fields)))

	_, err := db.Exec(query, values...)
	return err
}

func (mm *MessageModel) saveEmojiUsages(db *sql.DB, emojiModels []EmojiModel, gid string) error {
	for _, emj := range emojiModels {
		valuesMap := map[string]any{
			"message_id":  mm.Message.ID,
			"guild_id":    gid,
			"channel_id":  mm.Message.ChannelID,
			"author_id":   emj.AuthorID,
			"emoji_id":    emj.Emoji.ID,
			"is_reaction": emj.IsReaction,
			"timestamp":   emj.Timestamp,
			"m_author_id": emj.MessageAuthorId,
		}

		fields := utils.GetKeysFromMap(valuesMap)
		values := utils.GetValuesFromMapBasedOnKeys(valuesMap, fields)

		query := fmt.Sprintf(`
			INSERT INTO emoji_used (%s)
			VALUES (%s);
		`, strings.Join(fields, ", "), utils.SQLPlaceholders(len(fields)))

		if _, err := db.Exec(query, values...); err != nil {
			return err
		}
	}
	return nil
}
