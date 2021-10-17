package actions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidBotCommentForInvalidComment(t *testing.T) {
	assert.False(t, IsValidBotComment(""), "Undetected missing action prefix")
	assert.False(t, IsValidBotComment("/sq-bot invalid-command"), "Undetected invalid bot command")
	assert.False(t, IsValidBotComment("Some context with /sq-bot review within"), "Incorrect bot prefix detected inside random comment")
}

func TestIsValidBotCommentForValidComment(t *testing.T) {
	assert.True(t, IsValidBotComment("/sq-bot review"), "Correct bot comment not recognized")
}
