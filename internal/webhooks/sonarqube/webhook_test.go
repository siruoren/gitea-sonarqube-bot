package sonarqube

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWebhook(t *testing.T) {
	raw := []byte(`{ "serverUrl": "https://example.com/sonarqube", "taskId": "AXouyxDpizdp4B1K", "status": "SUCCESS", "analysedAt": "2021-05-21T12:12:07+0000", "revision": "f84442009c09b1adc278b6aa80a3853419f54007", "changedAt": "2021-05-21T12:12:07+0000", "project": { "key": "pr-bot", "name": "PR Bot", "url": "https://example.com/sonarqube/dashboard?id=pr-bot" }, "branch": { "name": "PR-1337", "type": "PULL_REQUEST", "isMain": false, "url": "https://example.com/sonarqube/dashboard?id=pr-bot&pullRequest=PR-1337" }, "qualityGate": { "name": "PR Bot", "status": "OK", "conditions": [ { "metric": "new_reliability_rating", "operator": "GREATER_THAN", "value": "1", "status": "OK", "errorThreshold": "1" }, { "metric": "new_security_rating", "operator": "GREATER_THAN", "value": "1", "status": "OK", "errorThreshold": "1" }, { "metric": "new_maintainability_rating", "operator": "GREATER_THAN", "value": "1", "status": "OK", "errorThreshold": "1" }, { "metric": "new_security_hotspots_reviewed", "operator": "LESS_THAN", "status": "NO_VALUE", "errorThreshold": "100" } ] }, "properties": {} }`)
	response, ok := New(raw)

	assert.NotNil(t, response)
	assert.True(t, ok)
}

func TestNewWebhookInvalidJSON(t *testing.T) {
	raw := []byte(`{ "serverUrl": ["invalid-server-url-content"] }`)
	_, ok := New(raw)

	assert.False(t, ok)
}

func TestGetPRIndex(t *testing.T) {
	raw := []byte(`{ "serverUrl": "https://example.com/sonarqube", "taskId": "AXouyxDpizdp4B1K", "status": "SUCCESS", "analysedAt": "2021-05-21T12:12:07+0000", "revision": "f84442009c09b1adc278b6aa80a3853419f54007", "changedAt": "2021-05-21T12:12:07+0000", "project": { "key": "pr-bot", "name": "PR Bot", "url": "https://example.com/sonarqube/dashboard?id=pr-bot" }, "branch": { "name": "PR-1337", "type": "PULL_REQUEST", "isMain": false, "url": "https://example.com/sonarqube/dashboard?id=pr-bot&pullRequest=PR-1337" }, "qualityGate": { "name": "PR Bot", "status": "OK", "conditions": [ { "metric": "new_reliability_rating", "operator": "GREATER_THAN", "value": "1", "status": "OK", "errorThreshold": "1" }, { "metric": "new_security_rating", "operator": "GREATER_THAN", "value": "1", "status": "OK", "errorThreshold": "1" }, { "metric": "new_maintainability_rating", "operator": "GREATER_THAN", "value": "1", "status": "OK", "errorThreshold": "1" }, { "metric": "new_security_hotspots_reviewed", "operator": "LESS_THAN", "status": "NO_VALUE", "errorThreshold": "100" } ] }, "properties": {} }`)
	w, _ := New(raw)

	idx, err := w.GetPRIndex()

	assert.Equal(t, 1337, idx)
	assert.Nil(t, err)
}

func TestGetPRIndexInvalidBranchName(t *testing.T) {
	raw := []byte(`{ "serverUrl": "https://example.com/sonarqube", "taskId": "AXouyxDpizdp4B1K", "status": "SUCCESS", "analysedAt": "2021-05-21T12:12:07+0000", "revision": "f84442009c09b1adc278b6aa80a3853419f54007", "changedAt": "2021-05-21T12:12:07+0000", "project": { "key": "pr-bot", "name": "PR Bot", "url": "https://example.com/sonarqube/dashboard?id=pr-bot" }, "branch": { "name": "invalid", "type": "PULL_REQUEST", "isMain": false, "url": "https://example.com/sonarqube/dashboard?id=pr-bot&pullRequest=PR-1337" }, "qualityGate": { "name": "PR Bot", "status": "OK", "conditions": [ { "metric": "new_reliability_rating", "operator": "GREATER_THAN", "value": "1", "status": "OK", "errorThreshold": "1" }, { "metric": "new_security_rating", "operator": "GREATER_THAN", "value": "1", "status": "OK", "errorThreshold": "1" }, { "metric": "new_maintainability_rating", "operator": "GREATER_THAN", "value": "1", "status": "OK", "errorThreshold": "1" }, { "metric": "new_security_hotspots_reviewed", "operator": "LESS_THAN", "status": "NO_VALUE", "errorThreshold": "100" } ] }, "properties": {} }`)
	w, _ := New(raw)

	idx, err := w.GetPRIndex()

	assert.Equal(t, 0, idx)
	assert.NotNil(t, err)
}
