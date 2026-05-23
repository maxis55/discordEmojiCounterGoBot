package emoji_processing

import (
	"database/sql"
	"emoji-counter/cache"
	dbpkg "emoji-counter/db"
	"emoji-counter/utils"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strings"
)

type AuthorModel struct {
	Author *discordgo.User
}

// Remember upserts the author row, skipping the DB write when the cache says we
// already saw this author recently.
func (am AuthorModel) Remember(db *sql.DB) error {
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

	_, err := dbpkg.Exec(db, query, values...)

	if err == nil {
		cache.RememberKey(fmt.Sprintf(cache.AUTHOR_KEY, am.Author.ID))
	}

	return err
}

// RememberAuthorByID inserts a minimal authors row for an author we don't have
// full info for (e.g., a reactor on the remove-event path). Does not write to
// the cache — we want a later Remember() call with full info to still run and
// populate username/global_name.
func RememberAuthorByID(authorID string, db *sql.DB) error {
	if cache.KeyExists(fmt.Sprintf(cache.AUTHOR_KEY, authorID)) {
		return nil
	}
	_, err := dbpkg.Exec(db,
		`INSERT INTO authors (author_id) VALUES ($1) ON CONFLICT (author_id) DO NOTHING`,
		authorID,
	)
	return err
}
