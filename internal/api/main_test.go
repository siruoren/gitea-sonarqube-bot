package api

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	giteaSdk "gitea-sonarqube-pr-bot/internal/clients/gitea"
	sqSdk "gitea-sonarqube-pr-bot/internal/clients/sonarqube"
	"gitea-sonarqube-pr-bot/internal/settings"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type SonarQubeHandlerMock struct {
	mock.Mock
}

func (h *SonarQubeHandlerMock) Handle(r *http.Request) (int, string) {
	h.Called(r)
	return http.StatusOK, "test-execution"
}

type GiteaHandlerMock struct {
	mock.Mock
}

func (h *GiteaHandlerMock) HandleSynchronize(r *http.Request) (int, string) {
	h.Called(r)
	return http.StatusOK, "test-execution"
}

func (h *GiteaHandlerMock) HandleComment(r *http.Request) (int, string) {
	h.Called(r)
	return http.StatusOK, "test-execution"
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

// SETUP: mute logs
func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	gin.DefaultWriter = ioutil.Discard
	gin.DefaultErrorWriter = ioutil.Discard
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}

func TestNonAPIRoutes(t *testing.T) {
	router := New(new(GiteaHandlerMock), new(SonarQubeHandlerMock))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/favicon.ico", nil)
	router.Engine.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/ping", nil)
	router.Engine.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSonarQubeAPIRouteMissingProjectHeader(t *testing.T) {
	router := New(new(GiteaHandlerMock), new(SonarQubeHandlerMock))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/hooks/sonarqube", bytes.NewBuffer([]byte(`{}`)))
	router.Engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSonarQubeAPIRouteProcessing(t *testing.T) {
	sonarQubeHandlerMock := new(SonarQubeHandlerMock)
	sonarQubeHandlerMock.On("Handle", mock.IsType(&http.Request{}))

	router := New(new(GiteaHandlerMock), sonarQubeHandlerMock)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/hooks/sonarqube", bytes.NewBuffer([]byte(`{}`)))
	req.Header.Add("X-SonarQube-Project", "gitea-sonarqube-bot")
	router.Engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	sonarQubeHandlerMock.AssertNumberOfCalls(t, "Handle", 1)
	sonarQubeHandlerMock.AssertExpectations(t)
}

func TestGiteaAPIRouteMissingEventHeader(t *testing.T) {
	router := New(new(GiteaHandlerMock), new(SonarQubeHandlerMock))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/hooks/gitea", bytes.NewBuffer([]byte(`{}`)))
	router.Engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGiteaAPIRouteSynchronizeProcessing(t *testing.T) {
	giteaHandlerMock := new(GiteaHandlerMock)
	giteaHandlerMock.On("HandleSynchronize", mock.Anything, mock.Anything).Return(nil)
	giteaHandlerMock.On("HandleComment", mock.Anything, mock.Anything).Maybe()

	router := New(giteaHandlerMock, new(SonarQubeHandlerMock))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/hooks/gitea", bytes.NewBuffer([]byte(`{}`)))
	req.Header.Add("X-Gitea-Event", "pull_request")
	router.Engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	giteaHandlerMock.AssertNumberOfCalls(t, "HandleSynchronize", 1)
	giteaHandlerMock.AssertNumberOfCalls(t, "HandleComment", 0)
	giteaHandlerMock.AssertExpectations(t)
}

func TestGiteaAPIRouteCommentProcessing(t *testing.T) {
	giteaHandlerMock := new(GiteaHandlerMock)
	giteaHandlerMock.On("HandleSynchronize", mock.Anything, mock.Anything).Maybe()
	giteaHandlerMock.On("HandleComment", mock.Anything, mock.Anything).Return(nil)

	router := New(giteaHandlerMock, new(SonarQubeHandlerMock))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/hooks/gitea", bytes.NewBuffer([]byte(`{}`)))
	req.Header.Add("X-Gitea-Event", "issue_comment")
	router.Engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	giteaHandlerMock.AssertNumberOfCalls(t, "HandleSynchronize", 0)
	giteaHandlerMock.AssertNumberOfCalls(t, "HandleComment", 1)
	giteaHandlerMock.AssertExpectations(t)
}

func TestGiteaAPIRouteUnknownEvent(t *testing.T) {
	giteaHandlerMock := new(GiteaHandlerMock)
	giteaHandlerMock.On("HandleSynchronize", mock.Anything, mock.Anything).Maybe()
	giteaHandlerMock.On("HandleComment", mock.Anything, mock.Anything).Maybe()

	router := New(giteaHandlerMock, new(SonarQubeHandlerMock))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/hooks/gitea", bytes.NewBuffer([]byte(`{}`)))
	req.Header.Add("X-Gitea-Event", "unknown")
	router.Engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	giteaHandlerMock.AssertNumberOfCalls(t, "HandleSynchronize", 0)
	giteaHandlerMock.AssertNumberOfCalls(t, "HandleComment", 0)
	giteaHandlerMock.AssertExpectations(t)
}
