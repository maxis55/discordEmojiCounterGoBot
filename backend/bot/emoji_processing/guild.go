package emoji_processing

import (
	"database/sql"
	"emoji-counter/utils"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strings"
)

type GuildModel struct {
	Guild *discordgo.Guild
}

func SaveGuildInfo(discord *discordgo.Session, gid string, db *sql.DB) {
	guild, err := discord.Guild(gid)
	if err != nil {
		fmt.Println(err.Error())
	}

	err = (&GuildModel{Guild: guild}).remember(db)

	if err != nil {
		fmt.Println(err.Error())
	}

	ejModels := make([]EmojiModel, 0, len(guild.Emojis))

	for _, emoji := range guild.Emojis {
		if emoji == nil {
			continue
		}
		ejModels = append(ejModels, EmojiModel{
			Emoji:   *emoji,
			GuildID: &gid,
		})

	}

	rememberGuildEmojis(ejModels, db)

	channels, err := discord.GuildChannels(gid)
	if err != nil {
		fmt.Println(err.Error())
	}

	for _, channel := range channels {
		err = (&ChannelModel{Channel: channel}).remember(db)
		if err != nil {
			fmt.Println(err.Error())
		}
	}

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
		ON CONFLICT (guild_id) DO UPDATE
		SET name=EXCLUDED.name, region=EXCLUDED.region, member_count=EXCLUDED.member_count, icon=EXCLUDED.icon, owner_id=EXCLUDED.owner_id;
	`, strings.Join(fields, ", "), utils.SQLPlaceholders(len(fields)))

	_, err := db.Exec(query, values...)
	return err
}

func rememberGuildEmojis(ejs []EmojiModel, db *sql.DB) {
	for _, emoji := range ejs {
		err := emoji.forceRemember(db)
		if err != nil {
			fmt.Println(err.Error())
		}
	}
}
