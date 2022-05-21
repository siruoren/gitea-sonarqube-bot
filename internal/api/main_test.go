package api

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	giteaSdk "gitea-sonarqube-pr-bot/internal/clients/gitea"
	sqSdk "gitea-sonarqube-pr-bot/internal/clients/sonarqube"
	"gitea-sonarqube-pr-bot/internal/settings"
	webhook "gitea-sonarqube-pr-bot/internal/webhooks/sonarqube"

	"github.com/stretchr/testify/mock"
)

// Default SDK mocking
type HandlerPartialMock struct {
	mock.Mock
}

func (h *HandlerPartialMock) fetchDetails(w *webhook.Webhook) {
	h.Called(w)
}

type GiteaSdkMock struct {
	mock.Mock
}

func (h *GiteaSdkMock) PostComment(_ settings.GiteaRepository, _ int, _ string) error {
	return nil
}

func (h *GiteaSdkMock) DetermineHEAD(_ settings.GiteaRepository, _ int64) (string, error) {
	return "", nil
}

func (h *GiteaSdkMock) UpdateStatus(_ settings.GiteaRepository, _ string, _ giteaSdk.StatusDetails) error {
	return nil
}

type SQSdkMock struct {
	mock.Mock
}

func (h *SQSdkMock) GetMeasures(project string, branch string) (*sqSdk.MeasuresResponse, error) {
	return &sqSdk.MeasuresResponse{}, nil
}

func (h *SQSdkMock) GetPullRequestUrl(project string, index int64) string {
	return ""
}

func (h *SQSdkMock) GetPullRequest(project string, index int64) (*sqSdk.PullRequest, error) {
	return &sqSdk.PullRequest{
		Status: struct {
			QualityGateStatus string "json:\"qualityGateStatus\""
		}{
			QualityGateStatus: "OK",
		},
	}, nil
}

func (h *SQSdkMock) ComposeGiteaComment(data *sqSdk.CommentComposeData) (string, error) {
	return "", nil
}

func defaultMockPreparation(h *HandlerPartialMock) {
	h.On("fetchDetails", mock.Anything).Return(nil)
}

// SETUP: mute logs
func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}
