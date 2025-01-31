package bot

import (
	"database/sql"
	"discordEmojiCounterBot/utils"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strings"
)

type MessageModel struct {
	Message   *discordgo.Message
	Reactions []EmojiModel
}

type AuthorModel struct {
	Author *discordgo.User
}

type GuildModel struct {
	Guild *discordgo.Guild
}
type ChannelModel struct {
	Channel *discordgo.Channel
}

func queryAllGuildChannels(db *sql.DB, gid string) ([]ChannelModel, error) {
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

func queryChannelById(db *sql.DB, cid string) (*ChannelModel, error) {
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

func (model *GuildModel) remember(db *sql.DB) error {
	valuesMap := map[string]any{
		"guild_id":          model.Guild.ID,
		"name":              model.Guild.Name,
		"system_channel_id": model.Guild.SystemChannelID,
		"region":            model.Guild.Region,
		"member_count":      model.Guild.MemberCount,
		"icon":              model.Guild.Icon,
		"joined_at":         model.Guild.JoinedAt,
		"owner_id":          model.Guild.OwnerID,
	}

	fields := utils.GetKeysFromMap(valuesMap)
	values := utils.GetValuesFromMapBasedOnKeys(valuesMap, fields)

	query := fmt.Sprintf(`
		INSERT INTO guilds (%s)
		VALUES (%s)
		ON CONFLICT (guild_id) DO NOTHING;
	`, strings.Join(fields, ", "), utils.SQLPlaceholders(len(fields)))

	_, err := db.Exec(query, values...)
	return err
}

func (cm *ChannelModel) remember(db *sql.DB) error {
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
		ON CONFLICT (channel_id) DO NOTHING;
	`, strings.Join(fields, ", "), utils.SQLPlaceholders(len(fields)))

	_, err := db.Exec(query, values...)

	return err
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

func (model AuthorModel) remember(db *sql.DB) error {
	valuesMap := map[string]any{
		"author_id":   model.Author.ID,
		"verified":    model.Author.Verified,
		"username":    model.Author.Username,
		"global_name": model.Author.GlobalName,
		"bot":         model.Author.Bot,
		"system":      model.Author.System,
		"mfa_enabled": model.Author.MFAEnabled,
	}

	fields := utils.GetKeysFromMap(valuesMap)
	values := utils.GetValuesFromMapBasedOnKeys(valuesMap, fields)

	query := fmt.Sprintf(`
		INSERT INTO authors (%s)
		VALUES (%s)
		ON CONFLICT (author_id) DO NOTHING;
	`, strings.Join(fields, ", "), utils.SQLPlaceholders(len(fields)))

	_, err := db.Exec(query, values...)
	return err
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
