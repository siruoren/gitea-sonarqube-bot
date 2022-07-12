package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func getRequestData() []byte {
	return []byte(`{"serverUrl":"https://example.com","status":"SUCCESS","analysedAt":"2022-05-15T16:45:31+0000","revision":"378080777919s07657a07f7a3e2d05dc75f64edd","changedAt":"2022-05-15T16:41:39+0000","project":{"key":"gitea-sonarqube-bot","name":"Gitea SonarQube Bot","url":"https://example.com/dashboard?id=gitea-sonarqube-bot"},"branch":{"name":"PR-1822","type":"PULL_REQUEST","isMain":false,"url":"https://example.com/dashboard?id=gitea-sonarqube-bot&pullRequest=PR-1822"},"qualityGate":{"name":"GiteaSonarQubeBot","status":"OK","conditions":[{"metric":"new_reliability_rating","operator":"GREATER_THAN","value":"1","status":"OK","errorThreshold":"1"},{"metric":"new_security_rating","operator":"GREATER_THAN","value":"1","status":"OK","errorThreshold":"1"},{"metric":"new_maintainability_rating","operator":"GREATER_THAN","value":"1","status":"OK","errorThreshold":"1"},{"metric":"new_security_hotspots_reviewed","operator":"LESS_THAN","status":"OK","errorThreshold":"100"}]},"properties":{"sonar.analysis.sqbot":"378080777919s07657a07f7a3e2d05dc75f64edd"}}`)
}

func TestIsValidWebhook(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		actual, _ := isValidWebhook(getRequestData(), "sonarqube-test-webhook-secret", "647f2395d30b1b7efcb58d9338be5b69c2addb54faf6bde6314a57ea28f45467", "test-component")
		assert.True(t, actual, "Expected successful webhook signature validation")
	})

	t.Run("Nothing configured or provided", func(t *testing.T) {
		actual, _ := isValidWebhook(getRequestData(), "", "", "test-component")
		assert.True(t, actual, "Webhook signature validation not skipped")
	})

	t.Run("Signature decoding error", func(t *testing.T) {
		actual, err := isValidWebhook(getRequestData(), "sonarqube-test-webhook-secret", "invalid-signature", "test-component")
		assert.False(t, actual)
		assert.EqualError(t, err, "Error decoding signature for test-component webhook.", "Undetected signature encoding error")
	})

	t.Run("Signature mismatch", func(t *testing.T) {
		actual, err := isValidWebhook(getRequestData(), "sonarqube-test-webhook-secret", "fde6a666b7a1a46c27efb1961c17b46b6cf7aa13db5560e5ac95e801a18a92f3", "test-component")
		assert.False(t, actual)
		assert.EqualError(t, err, "Signature header does not match the received test-component webhook content. Request rejected.", "Undetected signature mismatch")
	})

	t.Run("Empty secret configuration", func(t *testing.T) {
		actual, err := isValidWebhook(getRequestData(), "", "647f2395d30b1b7efcb58d9338be5b69c2addb54faf6bde6314a57ea28f45467", "test-component")
		assert.False(t, actual)
		assert.EqualError(t, err, "Signature header received but no test-component webhook secret configured. Request rejected due to possible configuration mismatch.", "Undetected configuration mismatch (1)")
	})

	t.Run("Empty signature configuration", func(t *testing.T) {
		actual, err := isValidWebhook(getRequestData(), "sonarqube-test-webhook-secret", "", "test-component")
		assert.False(t, actual)
		assert.EqualError(t, err, "test-component webhook secret configured but no signature header received. Request rejected due to possible configuration mismatch.", "Undetected configuration mismatch (2)")
	})
}
