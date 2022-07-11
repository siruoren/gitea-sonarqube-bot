package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"gitea-sonarqube-pr-bot/internal/settings"

	"github.com/stretchr/testify/assert"
)

func withValidSonarQubeRequestData(t *testing.T, jsonBody []byte) (*http.Request, *httptest.ResponseRecorder, http.HandlerFunc) {
	webhookHandler := NewSonarQubeWebhookHandler(new(GiteaSdkMock), new(SQSdkMock))

	req, err := http.NewRequest("POST", "/hooks/sonarqube", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-SonarQube-Project", "pr-bot")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status, response := webhookHandler.Handle(r)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		io.WriteString(w, fmt.Sprintf(`{"message": "%s"}`, response))
	})

	return req, rr, handler
}

func TestHandleSonarQubeWebhookProjectMapped(t *testing.T) {
	settings.Pattern = &settings.PatternConfig{
		RegExp: regexp.MustCompile(`^PR-(\d+)$`),
	}
	settings.SonarQube = settings.SonarQubeConfig{
		Webhook: &settings.Webhook{
			Secret: "",
		},
	}
	settings.Projects = []settings.Project{
		{
			SonarQube: struct{ Key string }{
				Key: "pr-bot",
			},
		},
	}
	req, rr, handler := withValidSonarQubeRequestData(t, []byte(`{ "serverUrl": "https://example.com/sonarqube", "taskId": "AXouyxDpizdp4B1K", "status": "SUCCESS", "analysedAt": "2021-05-21T12:12:07+0000", "revision": "f84442009c09b1adc278b6aa80a3853419f54007", "changedAt": "2021-05-21T12:12:07+0000", "project": { "key": "pr-bot", "name": "PR Bot", "url": "https://example.com/sonarqube/dashboard?id=pr-bot" }, "branch": { "name": "PR-1337", "type": "PULL_REQUEST", "isMain": false, "url": "https://example.com/sonarqube/dashboard?id=pr-bot&pullRequest=PR-1337" }, "qualityGate": { "name": "PR Bot", "status": "OK", "conditions": [ { "metric": "new_reliability_rating", "operator": "GREATER_THAN", "value": "1", "status": "OK", "errorThreshold": "1" }, { "metric": "new_security_rating", "operator": "GREATER_THAN", "value": "1", "status": "OK", "errorThreshold": "1" }, { "metric": "new_maintainability_rating", "operator": "GREATER_THAN", "value": "1", "status": "OK", "errorThreshold": "1" }, { "metric": "new_security_hotspots_reviewed", "operator": "LESS_THAN", "status": "NO_VALUE", "errorThreshold": "100" } ] }, "properties": {} }`))
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, `{"message": "Processing data. See bot logs for details."}`, rr.Body.String())

	t.Cleanup(func() {
		settings.Pattern = nil
	})
}

func TestHandleSonarQubeWebhookProjectNotMapped(t *testing.T) {
	settings.Projects = []settings.Project{
		{
			SonarQube: struct{ Key string }{
				Key: "another-project",
			},
		},
	}
	req, rr, handler := withValidSonarQubeRequestData(t, []byte(`{ "serverUrl": "https://example.com/sonarqube", "taskId": "AXouyxDpizdp4B1K", "status": "SUCCESS", "analysedAt": "2021-05-21T12:12:07+0000", "revision": "f84442009c09b1adc278b6aa80a3853419f54007", "changedAt": "2021-05-21T12:12:07+0000", "project": { "key": "pr-bot", "name": "PR Bot", "url": "https://example.com/sonarqube/dashboard?id=pr-bot" }, "branch": { "name": "PR-1337", "type": "PULL_REQUEST", "isMain": false, "url": "https://example.com/sonarqube/dashboard?id=pr-bot&pullRequest=PR-1337" }, "qualityGate": { "name": "PR Bot", "status": "OK", "conditions": [ { "metric": "new_reliability_rating", "operator": "GREATER_THAN", "value": "1", "status": "OK", "errorThreshold": "1" }, { "metric": "new_security_rating", "operator": "GREATER_THAN", "value": "1", "status": "OK", "errorThreshold": "1" }, { "metric": "new_maintainability_rating", "operator": "GREATER_THAN", "value": "1", "status": "OK", "errorThreshold": "1" }, { "metric": "new_security_hotspots_reviewed", "operator": "LESS_THAN", "status": "NO_VALUE", "errorThreshold": "100" } ] }, "properties": {} }`))
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, `{"message": "Project 'pr-bot' not in configured list. Request ignored."}`, rr.Body.String())
}

func TestHandleSonarQubeWebhookInvalidJSONBody(t *testing.T) {
	settings.Projects = []settings.Project{
		{
			SonarQube: struct{ Key string }{
				Key: "pr-bot",
			},
		},
	}

	req, rr, handler := withValidSonarQubeRequestData(t, []byte(`{ "serverUrl": ["invalid-server-url-content"] }`))
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	assert.Equal(t, `{"message": "Error parsing POST body."}`, rr.Body.String())
}

func TestHandleSonarQubeWebhookInvalidWebhookSignature(t *testing.T) {
	settings.SonarQube = settings.SonarQubeConfig{
		Webhook: &settings.Webhook{
			Secret: "sonarqube-test-webhook-secret",
		},
	}
	settings.Projects = []settings.Project{
		{
			SonarQube: struct{ Key string }{
				Key: "pr-bot",
			},
		},
	}
	req, rr, handler := withValidSonarQubeRequestData(t, []byte(`{ "serverUrl": "https://example.com/sonarqube", "taskId": "AXouyxDpizdp4B1K", "status": "SUCCESS", "analysedAt": "2021-05-21T12:12:07+0000", "revision": "f84442009c09b1adc278b6aa80a3853419f54007", "changedAt": "2021-05-21T12:12:07+0000", "project": { "key": "pr-bot", "name": "PR Bot", "url": "https://example.com/sonarqube/dashboard?id=pr-bot" }, "branch": { "name": "PR-1337", "type": "PULL_REQUEST", "isMain": false, "url": "https://example.com/sonarqube/dashboard?id=pr-bot&pullRequest=PR-1337" }, "qualityGate": { "name": "PR Bot", "status": "OK", "conditions": [ { "metric": "new_reliability_rating", "operator": "GREATER_THAN", "value": "1", "status": "OK", "errorThreshold": "1" }, { "metric": "new_security_rating", "operator": "GREATER_THAN", "value": "1", "status": "OK", "errorThreshold": "1" }, { "metric": "new_maintainability_rating", "operator": "GREATER_THAN", "value": "1", "status": "OK", "errorThreshold": "1" }, { "metric": "new_security_hotspots_reviewed", "operator": "LESS_THAN", "status": "NO_VALUE", "errorThreshold": "100" } ] }, "properties": {} }`))
	req.Header.Set("X-Sonar-Webhook-HMAC-SHA256", "647f2395d30b1b7efcb58d9338be5b69c2addb54faf6bde6314a57ea28f45467")
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusPreconditionFailed, rr.Code)
	assert.Equal(t, `{"message": "Webhook validation failed. Request rejected."}`, rr.Body.String())
}

func TestHandleSonarQubeWebhookForPullRequest(t *testing.T) {
	settings.Pattern = &settings.PatternConfig{
		RegExp: regexp.MustCompile(`^PR-(\d+)$`),
	}
	settings.SonarQube = settings.SonarQubeConfig{
		Webhook: &settings.Webhook{
			Secret: "",
		},
	}
	settings.Projects = []settings.Project{
		{
			SonarQube: struct{ Key string }{
				Key: "pr-bot",
			},
		},
	}

	req, rr, handler := withValidSonarQubeRequestData(t, []byte(`{ "serverUrl": "https://example.com/sonarqube", "taskId": "AXouyxDpizdp4B1K", "status": "SUCCESS", "analysedAt": "2021-05-21T12:12:07+0000", "revision": "f84442009c09b1adc278b6aa80a3853419f54007", "changedAt": "2021-05-21T12:12:07+0000", "project": { "key": "pr-bot", "name": "PR Bot", "url": "https://example.com/sonarqube/dashboard?id=pr-bot" }, "branch": { "name": "PR-1337", "type": "PULL_REQUEST", "isMain": false, "url": "https://example.com/sonarqube/dashboard?id=pr-bot&pullRequest=PR-1337" }, "qualityGate": { "name": "PR Bot", "status": "OK", "conditions": [ { "metric": "new_reliability_rating", "operator": "GREATER_THAN", "value": "1", "status": "OK", "errorThreshold": "1" }, { "metric": "new_security_rating", "operator": "GREATER_THAN", "value": "1", "status": "OK", "errorThreshold": "1" }, { "metric": "new_maintainability_rating", "operator": "GREATER_THAN", "value": "1", "status": "OK", "errorThreshold": "1" }, { "metric": "new_security_hotspots_reviewed", "operator": "LESS_THAN", "status": "NO_VALUE", "errorThreshold": "100" } ] }, "properties": {} }`))
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, `{"message": "Processing data. See bot logs for details."}`, rr.Body.String())

	t.Cleanup(func() {
		settings.Pattern = nil
	})
}

func TestHandleSonarQubeWebhookForBranch(t *testing.T) {
	settings.Pattern = &settings.PatternConfig{
		RegExp: regexp.MustCompile(`^PR-(\d+)$`),
	}
	settings.SonarQube = settings.SonarQubeConfig{
		Webhook: &settings.Webhook{
			Secret: "",
		},
	}
	settings.Projects = []settings.Project{
		{
			SonarQube: struct{ Key string }{
				Key: "pr-bot",
			},
		},
	}

	req, rr, handler := withValidSonarQubeRequestData(t, []byte(`{ "serverUrl": "https://example.com/sonarqube", "taskId": "AXouyxDpizdp4B1K", "status": "SUCCESS", "analysedAt": "2021-05-21T12:12:07+0000", "revision": "f84442009c09b1adc278b6aa80a3853419f54007", "changedAt": "2021-05-21T12:12:07+0000", "project": { "key": "pr-bot", "name": "PR Bot", "url": "https://example.com/sonarqube/dashboard?id=pr-bot" }, "branch": { "name": "PR-1337", "type": "BRANCH", "isMain": false, "url": "https://example.com/sonarqube/dashboard?id=pr-bot&pullRequest=PR-1337" }, "qualityGate": { "name": "PR Bot", "status": "OK", "conditions": [ { "metric": "new_reliability_rating", "operator": "GREATER_THAN", "value": "1", "status": "OK", "errorThreshold": "1" }, { "metric": "new_security_rating", "operator": "GREATER_THAN", "value": "1", "status": "OK", "errorThreshold": "1" }, { "metric": "new_maintainability_rating", "operator": "GREATER_THAN", "value": "1", "status": "OK", "errorThreshold": "1" }, { "metric": "new_security_hotspots_reviewed", "operator": "LESS_THAN", "status": "NO_VALUE", "errorThreshold": "100" } ] }, "properties": {} }`))
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, `{"message": "Ignore Hook for non-PR analysis."}`, rr.Body.String())

	t.Cleanup(func() {
		settings.Pattern = nil
	})
}
