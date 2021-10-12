package sonarqube

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"gitea-sonarqube-pr-bot/internal/actions"
	"gitea-sonarqube-pr-bot/internal/settings"
)

func ParsePRIndex(name string) (int, error) {
	re := regexp.MustCompile(`^PR-(\d+)$`)
	res := re.FindSubmatch([]byte(name))
	if len(res) != 2 {
		return 0, fmt.Errorf("branch name '%s' does not match regex '%s'", name, re.String())
	}

	return strconv.Atoi(string(res[1]))
}

func PRNameFromIndex(index int64) string {
	return fmt.Sprintf("PR-%d", index)
}

func GetRenderedQualityGate(qg string) string {
	status := ":white_check_mark:"
	if qg != "OK" {
		status = ":x:"
	}

	return fmt.Sprintf("**Quality Gate**: %s", status)
}

type SonarQubeSdkInterface interface {
	GetMeasures(string, string) (*MeasuresResponse, error)
	GetPullRequestUrl(string, int64) string
	GetPullRequest(string, int64) (*PullRequest, error)
	ComposeGiteaComment(*CommentComposeData) (string, error)
}

type CommentComposeData struct {
	Key         string
	PRName      string
	Url         string
	QualityGate string
}

type SonarQubeSdk struct {
	client  *http.Client
	baseUrl string
	token   string
}

func (sdk *SonarQubeSdk) GetPullRequestUrl(project string, index int64) string {
	return fmt.Sprintf("%s/dashboard?id=%s&pullRequest=%s", sdk.baseUrl, project, PRNameFromIndex(index))
}

func (sdk *SonarQubeSdk) fetchPullRequests(project string) *PullsResponse {
	url := fmt.Sprintf("%s/api/project_pull_requests/list?project=%s", sdk.baseUrl, project)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Printf("Cannot initialize Request: %s", err.Error())
		return nil
	}
	req.Header.Add("Authorization", sdk.basicAuth())
	rawResp, _ := sdk.client.Do(req)
	if rawResp.Body != nil {
		defer rawResp.Body.Close()
	}

	body, _ := io.ReadAll(rawResp.Body)
	response := &PullsResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Printf("cannot parse response from SonarQube: %s", err.Error())
		return nil
	}

	return response
}

func (sdk *SonarQubeSdk) GetPullRequest(project string, index int64) (*PullRequest, error) {
	response := sdk.fetchPullRequests(project)
	if response == nil {
		return nil, fmt.Errorf("unable to retrieve pull requests from SonarQube")
	}

	name := PRNameFromIndex(index)
	pr := response.GetPullRequest(name)
	if pr == nil {
		return nil, fmt.Errorf("no pull request found with name '%s'", name)
	}

	return pr, nil
}

func (sdk *SonarQubeSdk) GetMeasures(project string, branch string) (*MeasuresResponse, error) {
	url := fmt.Sprintf("%s/api/measures/component?additionalFields=metrics&metricKeys=%s&component=%s&pullRequest=%s", sdk.baseUrl, settings.SonarQube.GetMetricsList(), project, branch)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot initialize Request: %w", err)
	}
	req.Header.Add("Authorization", sdk.basicAuth())
	rawResp, _ := sdk.client.Do(req)
	if rawResp.Body != nil {
		defer rawResp.Body.Close()
	}

	body, _ := io.ReadAll(rawResp.Body)
	response := &MeasuresResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("cannot parse response from SonarQube: %w", err)
	}

	return response, nil
}

func (sdk *SonarQubeSdk) ComposeGiteaComment(data *CommentComposeData) (string, error) {
	m, err := sdk.GetMeasures(data.Key, data.PRName)
	if err != nil {
		log.Printf("Error composing Gitea comment: %s", err.Error())
		return "", err
	}

	message := make([]string, 5)
	message[0] = GetRenderedQualityGate(data.QualityGate)
	message[1] = m.GetRenderedMarkdownTable()
	message[2] = fmt.Sprintf(`See <a href="%s" target="_blank" rel="nofollow">SonarQube</a> for details.`, data.Url)
	message[3] = "---"
	message[4] = fmt.Sprintf("- If you want the bot to check again, post `%s`", actions.ActionReview)

	return strings.Join(message, "\n\n"), nil
}

func (sdk *SonarQubeSdk) basicAuth() string {
	auth := []byte(fmt.Sprintf("%s:", sdk.token))
	return fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString(auth))
}

func New() *SonarQubeSdk {
	return &SonarQubeSdk{
		client:  &http.Client{},
		baseUrl: settings.SonarQube.Url,
		token:   settings.SonarQube.Token.Value,
	}
}
