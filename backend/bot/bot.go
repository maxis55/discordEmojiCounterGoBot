package bot

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
)

var Token string
var dbv *sql.DB

var dancers = [...]string{"💃", "💃🏻", "💃🏼", "💃🏽", "💃🏾", "💃🏿", "🕺🏿", "🕺🏾", "🕺🏽", "🕺🏼", "🕺🏻", "🕺"}

type DiscordWebhookMessage struct {
	Content string `json:"content"`
}

func checkNilErr(e error) {
	if e != nil {
		log.Fatal("Error message")
	}
}

const pr = "%%"
const rankUsedEmojisInGuild = pr + "rankUsedEmojisInGuild"

func Run(dbc *sql.DB) {

	dbv = dbc

	// create a session
	discord, err := discordgo.New("Bot " + Token)
	checkNilErr(err)

	// add a event handler
	discord.AddHandler(newMessage)
	discord.AddHandler(messageUpdated)
	discord.AddHandler(newReaction)
	discord.AddHandler(removedReaction)
	discord.AddHandler(removedAllReactions)

	// open session
	err = discord.Open()

	NotifyAboutErrorViaWebhook(err)

	defer discord.Close() // close session, after function termination

	// keep bot running until there is NO os interruption (ctrl + C)
	fmt.Println("Bot running....")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

func newMessage(discord *discordgo.Session, message *discordgo.MessageCreate) {

	/* prevent bot responding to its own message
	this is achived by looking into the message author id
	if message.author.id is same as bot.author.id then just return
	*/
	if message.Author.ID == discord.State.User.ID || message.Author.Bot {
		return
	}

	// respond to user message if it contains `!help` or `!bye`
	switch {
	case strings.HasPrefix(message.Content, "%%hello"):
		discord.ChannelMessageSend(message.ChannelID, "Hello World😃")
	case strings.HasPrefix(message.Content, "%%bye"):
		discord.ChannelMessageSend(message.ChannelID, "Good Bye👋")

	case strings.HasPrefix(message.Content, "%%saveEverythingAboutThisGuild"):
		discord.ChannelMessageSend(message.ChannelID, "Ok")
		saveGuildInfo(discord, message.GuildID, dbv)
		discord.ChannelMessageSendReply(message.ChannelID, "Done", message.Reference())

	case strings.HasPrefix(message.Content, "%%danceInEveryChannel"):
		danceInEveryChannel(discord, message)

	case strings.HasPrefix(message.Content, "%%danceHere"):
		danceHere(discord, message)

	case strings.HasPrefix(message.Content, "%%helpMeRankEmojis"):
		discord.ChannelMessageSend(message.ChannelID, "This is an example, figure it out: %%rankUsedEmojisInGuild author=123 channel=123321 ignoreReactions=true belongToTheGuild=false ignoreMessageText=false fromDate=2022-01-01 toDate=2024-01-01 desc=true limit=10")
	case strings.HasPrefix(message.Content, "%%helpMeRankReactions"):
		discord.ChannelMessageSend(message.ChannelID, "This is a special case messageAuthor only works like this(dates are optional): %%rankUsedEmojisInGuild messageAuthor=123 ignoreMessageText=true fromDate=2022-01-01 toDate=2024-01-01 desc=true limit=10")
	case strings.HasPrefix(message.Content, rankUsedEmojisInGuild):
		s := ExtractSettings(message.Content)

		js, _ := json.Marshal(s)

		discord.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Rank with following settings: %s", js))

		rankedEmojis, err := getRankedUsedEmojisInGuild(dbv, message.GuildID, s)

		if err != nil {
			NotifyAboutErrorViaWebhook(err)
			return
		}

		embeds := rankedEmojis.TransformIntoDiscordEmbeds(string(js))

		for _, embed := range embeds {
			_, err = discord.ChannelMessageSendEmbed(message.ChannelID, &embed)

			NotifyAboutErrorViaWebhook(err)
		}
		return

		var fields []*discordgo.MessageEmbedField

		emojiRanking := []EmojiRanking{
			{1, "😀"},
			{2, "🎉"},
			{3, ":oldEmoji:"},
			{4, "😂"},
			{5, ":deletedEmoji:"},
			{6, "😎"},
			{7, "<:blusheze:825829907024052234>"},
			{8, "🐱"},
			{9, "🤔"},
			{10, "🔥"},
			{11, "💪"},
			{12, "🥳"},
			{13, "😅"},
			{14, "❤️"},
			{15, "🤩"},
			{16, "🥺"},
			{17, ":some very long name actually that doesnt mean anything but as well see what happens with it in the field:"},
			{18, "🎶"},
			{19, "💥"},
			{20, "✨"},
			{21, "😜"},
			{22, "💯"},
			{23, "🔥"},
			{24, "🙌"},
			{25, "👑"},
		}

		// Number of columns (8 columns)
		columns := 3
		perColumn := len(emojiRanking) / columns
		if len(emojiRanking)%columns != 0 {
			perColumn++
		}

		// Create content for each column
		columnContents := []string{}
		for i := 0; i < columns; i++ {
			startIndex := i * perColumn
			endIndex := (i + 1) * perColumn
			columnContents = append(columnContents, createColumnContent(startIndex, endIndex, emojiRanking))
			//fields = append(fields, &discordgo.MessageEmbedField{Name: "\u200B", Value: createColumnContent(startIndex, endIndex, emojiRanking), Inline: true})
		}
		//5500 spare symbols
		//lorem := "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean massa. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Donec quam felis, ultricies nec, pellentesque eu, pretium quis, sem. Nulla consequat massa quis enim. Donec pede justo, fringilla vel, aliquet nec, vulputate eget, arcu. In enim justo, rhoncus ut, imperdiet a, venenatis vitae, justo. Nullam dictum felis eu pede mollis pretium. Integer tincidunt. Cras dapibus. Vivamus elementum semper nisi. Aenean vulputate eleifend tellus. Aenean leo ligula, porttitor eu, consequat vitae, eleifend ac, enim. Aliquam lorem ante, dapibus in, viverra quis, feugiat a, tellus. Phasellus viverra nulla ut metus varius laoreet. Quisque rutrum. Aenean imperdiet. Etiam ultricies nisi vel augue. Curabitur ullamcorper ultricies nisi. Nam eget dui. Etiam rhoncus. Maecenas tempus, tellus eget condimentum rhoncus, sem quam semper libero, sit amet adipiscing sem neque sed ipsum. Nam quam nunc, blandit ve"
		//lorem2 := "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean massa. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Donec quam felis, ultricies nec, pellentesque eu, pretium quis, sem. Nulla consequat massa quis enim. Donec pede justo, fringilla vel, aliquet nec, vulputate eget, arcu. In enim justo, rhoncus ut, imperdiet a, venenatis vitae, justo. Nullam dictum felis eu pede mollis pretium. Integer tincidunt. Cras dapibus. Vivamus elementum semper nisi. Aenean vulputate eleifend tellus. Aenean leo ligula, porttitor eu, consequat vitae, eleifend ac, enim. Aliquam lorem ante, dapibus in, viverra quis, feugiat a, tellus. Phasellus viverra nulla ut metus varius laoreet. Quisque rutrum. Aenean imperdiet. Etiam ultricies nisi vel augue. Curabitur ullamcorper ultricies nisi. Nam eget dui. Etiam rhoncus. Maecenas tempus, tellus eget condimentum rhoncus, sem quam semper libero, sit amet adipiscing sem neque sed ipsum. Nam quam nunc, blandit vel, luctus pulvinar, hendrerit id, lorem. Maecenas nec odio et ante tincidunt tempus. Donec vitae sapien ut libero venenatis faucibus. Nullam quis ante. Etiam sit amet orci eget eros faucibus tincidunt. Duis leo. Sed fringilla mauris sit amet nibh. Donec sodales sagittis magna. Sed consequat, leo eget bibendum sodales, augue velit cursus nunc, quis gravida magna mi a libero. Fusce vulputate eleifend sapien. Vestibulum purus quam, scelerisque ut, mollis sed, nonummy id, metus. Nullam accumsan lorem in dui. Cras ultricies mi eu turpis hendrerit fringilla. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia Curae; In ac dui quis mi consectetuer lacinia. Nam pretium turpis et arcu. Duis arcu tortor, suscipit eget, imperdiet nec, imperdiet iaculis, ipsum. Sed aliquam ultrices mauris. Integer ante arcu, accumsan a, consectetuer eget, posuere ut, mauris. Praesent adipiscing. Phasellus ullamcorper ipsum rutrum nunc. Nunc nonummy metus. Vestibulum volutpat pretium libero. Cras id dui. Aenean ut eros et nisl sagittis vestibulum. Nullam nulla eros, ultricies sit amet, nonummy id, imperdiet feugiat, pede. Sed lectus. Donec mollis hendrerit risus. Phasellus nec sem in justo pellentesque facilisis. Etiam imperdiet imperdiet orci. Nunc nec neque. Phasellus leo dolor, tempus non, auctor et, hendrerit quis, nisi. Curabitur ligula sapien, tincidunt non, euismod vitae, posuere imperdiet, leo. Maecenas malesuada. Praesent congue erat at massa. Sed cursus turpis vitae tortor. Donec posuere vulputate arcu. Phasellus accumsan cursus velit. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia Curae; Sed aliquam, nisi quis porttitor congue, elit erat euismod orci, ac placerat dolor lectus quis orci. Phasellus consectetuer vestibulum elit. Aenean tellus metus, bibendum sed, posuere ac, mattis non, nunc. Vestibulum fringilla pede sit amet augue. In turpis. Pellentesque posuere. Praesent turpis. Aenean posuere, tortor sed cursus feugiat, nunc augue blandit nunc, eu sollicitudin urna dolor sagittis lacus. Donec elit libero, sodales nec, volutpat a, suscipit non, turpis. Nullam sagittis. Suspendisse pulvinar, augue ac venenatis condimentum, sem libero volutpat nibh, nec pellentesque velit pede quis nunc. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia Curae; Fusce id purus. Ut varius tincidunt libero. Phasellus dolor. Maecenas vestibulum mollis diam. Pellentesque ut neque. Pellentesque habitant morbi tristique senectus et netus et malesuada fames ac turpis egestas. In dui magna, posuere eget, vestibulum et, tempor auctor, justo. In ac felis quis tortor malesuada pretium. Pellentesque auctor neque nec urna. Proin sapien ipsum, porta a, auctor quis, euismod ut, mi. Aenean viverra rhoncus pede. Pellentesque habitant morbi tristique senectus et netus et malesuada fames ac turpis egestas. Ut non enim eleifend felis pretium feugiat. Vivamus quis mi. Phasellus a est. Phasellus magna. In hac habitasse platea dictumst. Curabitur at lacus ac velit ornare lobortis. Curabitur a felis in nunc fringilla tristique. Morbi mattis ullamcorper velit. Phasellus gravida semper nisi. Nullam vel sem. Pellentesque libero tortor, tincidunt et, tincidunt eget, semper nec, quam. Sed hendrerit. Morbi ac felis. Nunc egestas, augue at pellentesque laoreet, felis eros vehicula leo, at malesuada velit leo quis pede. Donec interdum, metus et hendrerit aliquet, dolor diam sagittis ligula, eget egestas libero turpis vel mi. Nunc nulla. Fusce risus nisl, viverra et, tempor et, pretium in, sapien. Donec venenatis vulputate lorem. Morbi nec metus. Phasellus blandit leo ut odio. Maecenas ullamcorper, dui et placerat feugiat, eros pede varius nisi, condimentum viverra felis nunc et lorem. Sed magna purus, fermentum eu, tincidunt eu, varius ut, felis. In auctor lobortis lacus. Quisque libero metus, condimentum nec, tempor a, commodo mollis, magna. Vestibulum ullamcorper mauris at ligula. Fusce fermentum. Nullam cursus lacinia erat. Praesent blandit laoreet nibh. Fusce convallis metus id felis luctus adipiscing. Pellentesque egestas, neque sit amet convallis pulvinar, justo nulla eleifend augue, ac auctor orci leo non est. Quisque id mi. Ut tincidunt tincidunt erat. Etiam feugiat lorem non metus. Vestibulum dapibus nunc ac augue. Curabitur vestibulum aliquam leo. Praesent egestas neque eu enim. In hac habitasse platea dictumst. Fusce a quam. Etiam ut purus mattis mauris sodale"
		//fields = append(fields, &discordgo.MessageEmbedField{Name: "\u200B", Value: lorem, Inline: true})
		//fields = append(fields, &discordgo.MessageEmbedField{Name: "\u200B", Value: lorem2, Inline: true})

		multileLineStr := "valu\n" + strings.Repeat("valu\n", 204)

		fields = append(fields, &discordgo.MessageEmbedField{Name: "\u200B", Value: multileLineStr, Inline: true})

		embed := discordgo.MessageEmbed{Title: "I hate my life", Fields: fields}

		//// Number of columns (8 columns)
		//columns := 8
		//perColumn := len(emojiRanking) / columns
		//if len(emojiRanking)%columns != 0 {
		//	perColumn++
		//}
		//
		//// Create content for each column
		//columnContents := []string{}
		//for i := 0; i < columns; i++ {
		//	startIndex := i * perColumn
		//	endIndex := (i + 1) * perColumn
		//	columnContents = append(columnContents, createColumnContent(startIndex, endIndex, emojiRanking))
		//	fields = append(fields, &discordgo.MessageEmbedField{Name: fmt.Sprintf("col%d", i+1), Value: createColumnContent(startIndex, endIndex, emojiRanking), Inline: true})
		//}

		//fields = append(fields, &discordgo.MessageEmbedField{Name: "col1", Value: column1, Inline: true})
		//fields = append(fields, &discordgo.MessageEmbedField{Name: "col2", Value: column2, Inline: true})
		//fields = append(fields, &discordgo.MessageEmbedField{Name: "col3", Value: column3, Inline: true})

		fmt.Println(rankedEmojis)
		_, err = discord.ChannelMessageSendEmbedReply(message.ChannelID, &embed, message.Reference())

		NotifyAboutErrorViaWebhook(err)

		//discord.ChannelMessageSendReply(message.ChannelID, res, message.Reference())
	}

	err := ProcessOneMessage(nil, MessageModel{Message: message.Message}, message.GuildID, dbv, true)

	NotifyAboutErrorViaWebhook(err)

	return
}

type EmojiRanking struct {
	Rank  int
	Emoji string
}

func createColumnContent(startIndex, endIndex int, ranking []EmojiRanking) string {
	var column []string
	for i := startIndex; i < endIndex && i < len(ranking); i++ {
		column = append(column, fmt.Sprintf("**%d.** %s", ranking[i].Rank, ranking[i].Emoji))
	}
	return strings.Join(column, "\n")
}

func messageUpdated(discord *discordgo.Session, message *discordgo.MessageUpdate) {
	err := ProcessOneMessage(nil, MessageModel{Message: message.Message}, message.GuildID, dbv, true)

	NotifyAboutErrorViaWebhook(err)
}

func newReaction(discord *discordgo.Session, messageReaction *discordgo.MessageReactionAdd) {
	processReaction(discord, messageReaction.MessageReaction)
}

func removedReaction(discord *discordgo.Session, messageReaction *discordgo.MessageReactionRemove) {
	processReaction(discord, messageReaction.MessageReaction)
}

func removedAllReactions(discord *discordgo.Session, messageReaction *discordgo.MessageReactionRemoveAll) {
	processReaction(discord, messageReaction.MessageReaction)
}

func processReaction(discord *discordgo.Session, messageReaction *discordgo.MessageReaction) {
	// ignore your own reactions just in case
	if messageReaction.UserID == discord.State.User.ID {
		return
	}

	//We don't know if it's a reaction under the bot message or not, reactions under the bot messages are ignored
	//so we need to re-fetch it

	msg, err := discord.ChannelMessage(messageReaction.ChannelID, messageReaction.MessageID, requestConfig)
	msg.GuildID = messageReaction.GuildID

	if err != nil {
		NotifyAboutErrorViaWebhook(err)
		return
	}

	err = ProcessOneMessage(discord, MessageModel{Message: msg}, messageReaction.GuildID, dbv, true)

	NotifyAboutErrorViaWebhook(err)
}

func NotifyAboutErrorViaWebhook(botErr error) {
	if botErr == nil {
		return
	}

	payload := DiscordWebhookMessage{
		Content: fmt.Sprintf("<@%s> 💀 Reason: %s", os.Getenv("BOT_MASTER_ID"), botErr.Error()),
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Println(fmt.Sprintf("Failed to marshal payload: %v", err))
		return
	}

	resp, err := http.Post(os.Getenv("DISCORD_WEBHOOK_URL"), "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Println(fmt.Sprintf("Failed to send webhook: %v", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		log.Println(fmt.Sprintf("Unexpected response from Discord: %s", resp.Status))
		return
	}
}
