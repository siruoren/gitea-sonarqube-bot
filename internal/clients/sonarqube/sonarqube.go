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

type SonarQubeSdkInterface interface {
	GetMeasures(string, string) (*MeasuresResponse, error)
	GetPullRequestUrl(string, int64) string
	GetPullRequest(string, int64) (*PullRequest, error)
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
	url := fmt.Sprintf("%s/api/measures/component?additionalFields=metrics&metricKeys=bugs,vulnerabilities,new_security_hotspots,code_smells&component=%s&pullRequest=%s", sdk.baseUrl, project, branch)
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
