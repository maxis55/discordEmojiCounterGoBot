package bot

import (
	"database/sql"
	"errors"
	"github.com/bwmarrin/discordgo"
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

func (model GuildModel) remember(db *sql.DB) error {

	params := []interface{}{
		model.Guild.ID,
		model.Guild.Name,
		model.Guild.SystemChannelID,
		model.Guild.Region,
		model.Guild.MemberCount,
		model.Guild.Icon,
		model.Guild.JoinedAt,
		model.Guild.OwnerID,
	}

	if _, err := db.Exec(`INSERT INTO guilds (guild_id, name, system_channel_id, region, member_count, icon, joined_at, owner_id)
									VALUES ($1, $2, $3, $4, $5, $6, $7, $8) ON CONFLICT (guild_id) DO NOTHING;`, params...); err != nil {
		return err
	}
	return nil
}

func (model ChannelModel) remember(db *sql.DB) error {
	if _, err := db.Exec("INSERT INTO channels (channel_id, owner_id, name, type, application_id, parent_id, guild_id, nsfw, position)"+
		" VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) ON CONFLICT (channel_id) DO NOTHING;",
		model.Channel.ID, model.Channel.OwnerID, model.Channel.Name, model.Channel.Type, model.Channel.ApplicationID, model.Channel.ParentID, model.Channel.GuildID, model.Channel.NSFW, model.Channel.Position); err != nil {
		return err
	}
	return nil
}

func (model MessageModel) remember(gid string, db *sql.DB) error {
	if _, err := db.Exec("INSERT INTO messages (message_id, content, guild_id, channel_id, type, author_id, timestamp, edited_timestamp)"+
		" VALUES ($1, $2, $3, $4, $5, $6, $7, $8) ON CONFLICT (message_id) DO NOTHING;",
		model.Message.ID, model.Message.Content, gid, model.Message.ChannelID, model.Message.Type, model.Message.Author.ID, model.Message.Timestamp, model.Message.EditedTimestamp); err != nil {
		return err
	}
	return nil
}

// guild id is required separately because message is not guaranteed to have it
func (model MessageModel) saveEmojiUsages(db *sql.DB, emojiModels []EmojiModel, gid string) error {

	for _, emj := range emojiModels {
		if _, err := db.Exec(`INSERT INTO emoji_used (message_id, guild_id, channel_id,  author_id, emoji_id, is_reaction)
			 VALUES ($1, $2, $3, $4, $5, $6);`,
			model.Message.ID, gid, model.Message.ChannelID, emj.AuthorID, emj.Emoji.ID, emj.IsReaction); err != nil {
			return err
		}
	}
	return nil
}

func (model AuthorModel) remember(db *sql.DB) error {
	if _, err := db.Exec("INSERT INTO authors (author_id, verified, username, global_name, bot, system, mfa_enabled)"+
		" VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT (author_id) DO NOTHING;",
		model.Author.ID, model.Author.Verified, model.Author.Username, model.Author.GlobalName, model.Author.Bot, model.Author.System, model.Author.MFAEnabled); err != nil {
		return err
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
