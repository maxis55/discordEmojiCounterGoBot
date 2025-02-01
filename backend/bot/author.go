package bot

import (
	"database/sql"
	"emoji-counter/cache"
	"emoji-counter/utils"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strings"
)

type AuthorModel struct {
	Author *discordgo.User
}

func (am AuthorModel) remember(db *sql.DB) error {
	if cache.KeyExists(fmt.Sprintf(cache.AUTHOR_KEY, am.Author.ID)) {
		return nil
	}

	valuesMap := map[string]any{
		"author_id":   am.Author.ID,
		"verified":    am.Author.Verified,
		"username":    am.Author.Username,
		"global_name": am.Author.GlobalName,
		"bot":         am.Author.Bot,
		"system":      am.Author.System,
		"mfa_enabled": am.Author.MFAEnabled,
	}

	fields := utils.GetKeysFromMap(valuesMap)
	values := utils.GetValuesFromMapBasedOnKeys(valuesMap, fields)

	query := fmt.Sprintf(`
		INSERT INTO authors (%s)
		VALUES (%s)
		ON CONFLICT (author_id) DO UPDATE
		SET username=EXCLUDED.username, global_name=EXCLUDED.global_name;
	`, strings.Join(fields, ", "), utils.SQLPlaceholders(len(fields)))

	_, err := db.Exec(query, values...)

	if err == nil {
		cache.RememberKey(fmt.Sprintf(cache.AUTHOR_KEY, am.Author.ID))
	}

	return err
}
