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
	valuesMap := map[string]any{
		"message_id":       mm.Message.ID,
		"content":          mm.Message.Content,
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
		ON CONFLICT (message_id) DO NOTHING;
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

func cleanInfoAboutMessage(mid string, db *sql.DB) error {
	if _, err := db.Exec("DELETE FROM messages WHERE message_id=$1;", mid); err != nil {
		return err
	}
	if _, err := db.Exec("DELETE FROM emoji_used where message_id=$1;", mid); err != nil {
		return err
	}
	return nil
}
