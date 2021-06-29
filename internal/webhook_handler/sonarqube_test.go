package webhook_handler

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitea-sonarqube-pr-bot/internal/settings"
	webhook "gitea-sonarqube-pr-bot/internal/webhooks/sonarqube"
)

func withValidRequestData(t *testing.T) (*http.Request, *httptest.ResponseRecorder, http.HandlerFunc) {
	webhookHandler := NewSonarQubeWebhookHandler()
	webhookHandler.fetchDetails = func(w *webhook.Webhook) {
		log.Printf("Overridden fetchDetails")
	}

	jsonBody := []byte(`{ "serverUrl": "https://example.com/sonarqube", "taskId": "AXouyxDpizdp4B1K", "status": "SUCCESS", "analysedAt": "2021-05-21T12:12:07+0000", "revision": "f84442009c09b1adc278b6aa80a3853419f54007", "changedAt": "2021-05-21T12:12:07+0000", "project": { "key": "pr-bot", "name": "PR Bot", "url": "https://example.com/sonarqube/dashboard?id=pr-bot" }, "branch": { "name": "PR-1337", "type": "PULL_REQUEST", "isMain": false, "url": "https://example.com/sonarqube/dashboard?id=pr-bot&pullRequest=PR-1337" }, "qualityGate": { "name": "PR Bot", "status": "OK", "conditions": [ { "metric": "new_reliability_rating", "operator": "GREATER_THAN", "value": "1", "status": "OK", "errorThreshold": "1" }, { "metric": "new_security_rating", "operator": "GREATER_THAN", "value": "1", "status": "OK", "errorThreshold": "1" }, { "metric": "new_maintainability_rating", "operator": "GREATER_THAN", "value": "1", "status": "OK", "errorThreshold": "1" }, { "metric": "new_security_hotspots_reviewed", "operator": "LESS_THAN", "status": "NO_VALUE", "errorThreshold": "100" } ] }, "properties": {} }`)
	req, err := http.NewRequest("POST", "/hooks/sonarqube", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-SonarQube-Project", "pr-bot")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(webhookHandler.Handle)

	return req, rr, handler
}

func TestHandleSonarQubeWebhookProjectMapped(t *testing.T) {
	settings.Projects = []settings.Project{
		settings.Project{
			SonarQube: struct{Key string}{
				Key: "pr-bot",
			},
		},
	}
	req, rr, handler := withValidRequestData(t)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, `{"message": "Processing data. See bot logs for details."}`, rr.Body.String())
}

func TestHandleSonarQubeWebhookProjectNotMapped(t *testing.T) {
	settings.Projects = []settings.Project{
		settings.Project{
			SonarQube: struct{Key string}{
				Key: "another-project",
			},
		},
	}
	req, rr, handler := withValidRequestData(t)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, `{"message": "Project 'pr-bot' not in configured list. Request ignored."}`, rr.Body.String())
}

func TestHandleSonarQubeWebhookInvalidJSONBody(t *testing.T) {
	settings.Projects = []settings.Project{
		settings.Project{
			SonarQube: struct{Key string}{
				Key: "pr-bot",
			},
		},
	}

	jsonBody := []byte(`{ "serverUrl": ["invalid-server-url-content"] }`)
	req, err := http.NewRequest("POST", "/hooks/sonarqube", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-SonarQube-Project", "pr-bot")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(NewSonarQubeWebhookHandler().Handle)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	assert.Equal(t, `{"message": "Error parsing POST body."}`, rr.Body.String())
}
