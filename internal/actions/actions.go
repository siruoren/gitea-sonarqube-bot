package actions

import "strings"

type BotAction string

const (
	ActionReview BotAction = "/sq-bot review"
	ActionPrefix string    = "/sq-bot"
)

func IsValidBotComment(c string) bool {
	if !strings.HasPrefix(c, ActionPrefix) {
		return false
	}

	if c != string(ActionReview) {
		return false
	}

	return true
}
