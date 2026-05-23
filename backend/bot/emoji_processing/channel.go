package emoji_processing

import (
	"database/sql"
	"emoji-counter/cache"
	dbpkg "emoji-counter/db"
	"emoji-counter/utils"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strings"
)

type ChannelModel struct {
	Channel *discordgo.Channel
}

func QueryAllGuildChannels(db *sql.DB, gid string) ([]ChannelModel, error) {
	rows, err := db.Query("SELECT channel_id, name FROM channels where guild_id=$1", gid)
	if err != nil {
		return nil, err
	}

	var channels []ChannelModel

	for rows.Next() {
		model := ChannelModel{Channel: &discordgo.Channel{}}
		err = rows.Scan(&model.Channel.ID, &model.Channel.Name)
		if err != nil {
			return nil, err
		}
		channels = append(channels, model)
	}

	return channels, nil
}

func QueryChannelById(db *sql.DB, cid string) (*ChannelModel, error) {
	row := db.QueryRow("SELECT channel_id, name FROM channels where channel_id=$1 LIMIT 1", cid)

	model := &ChannelModel{Channel: &discordgo.Channel{}}

	err := row.Scan(&model.Channel.ID, &model.Channel.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return model, nil
}

func (cm *ChannelModel) remember(db *sql.DB) error {
	if cache.KeyExists(fmt.Sprintf(cache.CHANNEL_KEY, cm.Channel.ID)) {
		return nil
	}

	valuesMap := map[string]any{
		"channel_id":     cm.Channel.ID,
		"owner_id":       cm.Channel.OwnerID,
		"name":           cm.Channel.Name,
		"type":           cm.Channel.Type,
		"application_id": cm.Channel.ApplicationID,
		"parent_id":      cm.Channel.ParentID,
		"guild_id":       cm.Channel.GuildID,
		"nsfw":           cm.Channel.NSFW,
		"position":       cm.Channel.Position,
	}

	fields := utils.GetKeysFromMap(valuesMap)
	values := utils.GetValuesFromMapBasedOnKeys(valuesMap, fields)

	query := fmt.Sprintf(`
		INSERT INTO channels (%s)
		VALUES (%s)
		ON CONFLICT (channel_id) DO UPDATE
		SET name=EXCLUDED.name, position=EXCLUDED.position, nsfw=EXCLUDED.nsfw;
	`, strings.Join(fields, ", "), utils.SQLPlaceholders(len(fields)))

	_, err := dbpkg.Exec(db, query, values...)

	if err == nil {
		cache.RememberKey(fmt.Sprintf(cache.CHANNEL_KEY, cm.Channel.ID))
	}

	return err
}
