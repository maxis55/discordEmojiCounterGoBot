package bot

import (
	"regexp"
	"strconv"
)

type RankingSettings struct {
	AuthorID           *string `json:"author,omitempty"`
	ChannelID          *string `json:"channel,omitempty"`
	WithoutReactions   *bool   `json:"withoutReactions,omitempty"`
	WithoutMessageText *bool   `json:"withoutMessageText,omitempty"`
	BelongToTheGuild   *bool   `json:"belongToTheGuild,omitempty"`
	FromDate           *string `json:"fromDate,omitempty"`
	ToDate             *string `json:"toDate,omitempty"`
	Desc               *bool   `json:"desc,omitempty"`
	Limit              *int    `json:"limit,omitempty"`
}

// ExtractSettings takes a string of parameters and returns a RankingSettings struct
func ExtractSettings(text string) RankingSettings {
	settings := RankingSettings{}

	// Define regex patterns for each parameter
	patterns := map[string]*regexp.Regexp{
		"author":             regexp.MustCompile(`author=(\d+)`),
		"withoutReactions":   regexp.MustCompile(`withoutReactions=(true|false)`),
		"withoutMessageText": regexp.MustCompile(`withoutMessageText=(true|false)`),
		"belongToTheGuild":   regexp.MustCompile(`belongToTheGuild=(true|false)`),
		"fromDate":           regexp.MustCompile(`fromDate=(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})`),
		"toDate":             regexp.MustCompile(`toDate=(\d{4}-\d{2}-\d{2})`),
		"desc":               regexp.MustCompile(`desc=(true|false)`),
		"limit":              regexp.MustCompile(`limit=(\d+)`),
		"channel":            regexp.MustCompile(`channel=(\d+)`),
	}

	// Apply each regex pattern to the text and assign to the struct
	for key, pattern := range patterns {
		if match := pattern.FindStringSubmatch(text); match != nil {
			switch key {
			case "author":
				settings.AuthorID = &match[1]
			case "channel":
				settings.ChannelID = &match[1]
			case "withoutReactions":
				b := match[1] == "true"
				settings.WithoutReactions = &b
			case "withoutMessageText":
				b := match[1] == "true"
				settings.WithoutMessageText = &b
			case "belongToTheGuild":
				b := match[1] == "true"
				settings.BelongToTheGuild = &b
			case "fromDate":
				settings.FromDate = &match[1]
			case "toDate":
				settings.ToDate = &match[1]
			case "desc":
				b := match[1] == "true"
				settings.Desc = &b
			case "limit":
				val, _ := strconv.Atoi(match[1])
				settings.Limit = &val
			}
		}
	}

	if settings.Limit == nil || *settings.Limit < 1 {
		l := 10
		settings.Limit = &l
	}

	if settings.Desc == nil {
		b := true
		settings.Desc = &b
	}

	return settings
}
