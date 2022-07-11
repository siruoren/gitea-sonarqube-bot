package sonarqube

import (
	"regexp"
	"testing"

	"gitea-sonarqube-bot/internal/settings"

	"github.com/stretchr/testify/assert"
)

func TestNewWebhook(t *testing.T) {
	settings.Pattern = &settings.PatternConfig{
		RegExp: regexp.MustCompile(`^PR-(\d+)$`),
	}

	raw := []byte(`{ "serverUrl": "https://example.com/sonarqube", "taskId": "AXouyxDpizdp4B1K", "status": "SUCCESS", "analysedAt": "2021-05-21T12:12:07+0000", "revision": "f84442009c09b1adc278b6aa80a3853419f54007", "changedAt": "2021-05-21T12:12:07+0000", "project": { "key": "pr-bot", "name": "PR Bot", "url": "https://example.com/sonarqube/dashboard?id=pr-bot" }, "branch": { "name": "PR-1337", "type": "PULL_REQUEST", "isMain": false, "url": "https://example.com/sonarqube/dashboard?id=pr-bot&pullRequest=PR-1337" }, "qualityGate": { "name": "PR Bot", "status": "OK", "conditions": [ { "metric": "new_reliability_rating", "operator": "GREATER_THAN", "value": "1", "status": "OK", "errorThreshold": "1" }, { "metric": "new_security_rating", "operator": "GREATER_THAN", "value": "1", "status": "OK", "errorThreshold": "1" }, { "metric": "new_maintainability_rating", "operator": "GREATER_THAN", "value": "1", "status": "OK", "errorThreshold": "1" }, { "metric": "new_security_hotspots_reviewed", "operator": "LESS_THAN", "status": "NO_VALUE", "errorThreshold": "100" } ] }, "properties": { "sonar.analysis.sqbot": "a84442009c09b1adc278b6bb80a3853419f54007" } }`)
	response, ok := New(raw)

	assert.NotNil(t, response)
	assert.Equal(t, 1337, response.PRIndex)
	assert.Equal(t, "a84442009c09b1adc278b6bb80a3853419f54007", response.Properties.OriginalCommit)
	assert.True(t, ok)

	t.Cleanup(func() {
		settings.Pattern = nil
	})
}

func TestNewWebhookInvalidJSON(t *testing.T) {
	raw := []byte(`{ "serverUrl": ["invalid-server-url-content"] }`)
	_, ok := New(raw)

	assert.False(t, ok)
}

func TestNewWebhookInvalidBranchName(t *testing.T) {
	settings.Pattern = &settings.PatternConfig{
		RegExp: regexp.MustCompile(`^PR-(\d+)$`),
	}

	raw := []byte(`{ "serverUrl": "https://example.com/sonarqube", "taskId": "AXouyxDpizdp4B1K", "status": "SUCCESS", "analysedAt": "2021-05-21T12:12:07+0000", "revision": "f84442009c09b1adc278b6aa80a3853419f54007", "changedAt": "2021-05-21T12:12:07+0000", "project": { "key": "pr-bot", "name": "PR Bot", "url": "https://example.com/sonarqube/dashboard?id=pr-bot" }, "branch": { "name": "invalid", "type": "PULL_REQUEST", "isMain": false, "url": "https://example.com/sonarqube/dashboard?id=pr-bot&pullRequest=PR-1337" }, "qualityGate": { "name": "PR Bot", "status": "OK", "conditions": [ { "metric": "new_reliability_rating", "operator": "GREATER_THAN", "value": "1", "status": "OK", "errorThreshold": "1" }, { "metric": "new_security_rating", "operator": "GREATER_THAN", "value": "1", "status": "OK", "errorThreshold": "1" }, { "metric": "new_maintainability_rating", "operator": "GREATER_THAN", "value": "1", "status": "OK", "errorThreshold": "1" }, { "metric": "new_security_hotspots_reviewed", "operator": "LESS_THAN", "status": "NO_VALUE", "errorThreshold": "100" } ] }, "properties": {} }`)
	_, ok := New(raw)

	assert.False(t, ok)

	t.Cleanup(func() {
		settings.Pattern = nil
	})
}

func TestWebhookGetRevision(t *testing.T) {
	t.Run("Default revision", func(t *testing.T) {
		w := Webhook{
			Revision: "225fa0306c0ab83297d0cb5db0717b194ccb2e76",
		}

		assert.Equal(t, w.Revision, w.GetRevision())
	})

	t.Run("Default revision due to incomplete properties", func(t *testing.T) {
		w := Webhook{
			Revision:   "225fa0306c0ab83297d0cb5db0717b194ccb2e76",
			Properties: &properties{},
		}

		assert.Equal(t, w.Revision, w.GetRevision())
	})

	t.Run("Original commit from properties", func(t *testing.T) {
		w := Webhook{
			Revision: "225fa0306c0ab83297d0cb5db0717b194ccb2e76",
			Properties: &properties{
				OriginalCommit: "a9fe6800b0bbb70748aff53a011d8c09bbff42fe",
			},
		}

		assert.Equal(t, w.Properties.OriginalCommit, w.GetRevision())
	})
}
