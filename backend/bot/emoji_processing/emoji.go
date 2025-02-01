package emoji_processing

import (
	"database/sql"
	"emoji-counter/cache"
	"emoji-counter/utils"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strings"
	"time"
)

type EmojiModel struct {
	Emoji           discordgo.Emoji
	AuthorID        *string
	GuildID         *string
	Timestamp       time.Time
	MessageAuthorId *string
	IsReaction      bool
}

func (em *EmojiModel) remember(db *sql.DB) error {
	if cache.KeyExists(fmt.Sprintf(cache.EMOJI_KEY, em.Emoji.ID)) {
		return nil
	}

	valuesMap := map[string]any{
		"emoji_id": em.Emoji.ID,
		"name":     em.Emoji.Name,
		"guild_id": em.GuildID,
		"animated": em.Emoji.Animated,
	}

	fields := utils.GetKeysFromMap(valuesMap)

	values := utils.GetValuesFromMapBasedOnKeys(valuesMap, fields)

	query := fmt.Sprintf(`
		INSERT INTO emojis (%s)
		VALUES (%s)
		ON CONFLICT (emoji_id) DO NOTHING;
	`, strings.Join(fields, ", "), utils.SQLPlaceholders(len(fields)))

	_, err := db.Exec(query, values...)

	if err == nil {
		cache.RememberKey(fmt.Sprintf(cache.EMOJI_KEY, em.Emoji.ID))
	}

	return err
}
