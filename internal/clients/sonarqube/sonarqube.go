package sonarqube

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"gitea-sonarqube-pr-bot/internal/actions"
	"gitea-sonarqube-pr-bot/internal/settings"
)

func ParsePRIndex(name string) (int, error) {
	res := settings.Pattern.RegExp.FindSubmatch([]byte(name))
	if len(res) != 2 {
		return 0, fmt.Errorf("branch name '%s' does not match regex '%s'", name, settings.Pattern.RegExp.String())
	}

	return strconv.Atoi(string(res[1]))
}

func PRNameFromIndex(index int64) string {
	return fmt.Sprintf(settings.Pattern.Template, index)
}

func GetRenderedQualityGate(qg string) string {
	status := ":white_check_mark:"
	if qg != "OK" {
		status = ":x:"
	}

	return fmt.Sprintf("**Quality Gate**: %s", status)
}

func retrieveDataFromApi(sdk *SonarQubeSdk, request *http.Request, wrapper interface{}) error {
	request.Header.Add("Authorization", sdk.basicAuth())
	rawResponse, err := sdk.client.Do(request)
	if err != nil {
		return err
	}

	if rawResponse.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("missing or invalid API token")
	}

	if rawResponse.Body != nil {
		defer rawResponse.Body.Close()
	}

	body, err := sdk.bodyReader(rawResponse.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, wrapper)
	if err != nil {
		return err
	}

	return nil
}

type Error struct {
	Message string `json:"msg"`
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

type ClientInterface interface {
	Do(req *http.Request) (*http.Response, error)
}

type BodyReader func(io.Reader) ([]byte, error)
type HttpRequest func(method string, target string, body io.Reader) (*http.Request, error)

type SonarQubeSdk struct {
	client      ClientInterface
	bodyReader  BodyReader
	httpRequest HttpRequest
	baseUrl     string
	token       string
}

func (sdk *SonarQubeSdk) GetPullRequestUrl(project string, index int64) string {
	return fmt.Sprintf("%s/dashboard?id=%s&pullRequest=%s", sdk.baseUrl, project, PRNameFromIndex(index))
}

func (sdk *SonarQubeSdk) fetchPullRequests(project string) (*PullsResponse, error) {
	url := fmt.Sprintf("%s/api/project_pull_requests/list?project=%s", sdk.baseUrl, project)
	request, err := sdk.httpRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	response := &PullsResponse{}
	err = retrieveDataFromApi(sdk, request, response)
	if err != nil {
		return nil, err
	}

	if len(response.Errors) != 0 {
		return nil, fmt.Errorf("%s", response.Errors[0].Message)
	}

	return response, nil
}

func (sdk *SonarQubeSdk) GetPullRequest(project string, index int64) (*PullRequest, error) {
	response, err := sdk.fetchPullRequests(project)
	if err != nil {
		return nil, fmt.Errorf("fetching pull requests failed: %w", err)
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
	request, err := sdk.httpRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	response := &MeasuresResponse{}
	err = retrieveDataFromApi(sdk, request, response)
	if err != nil {
		return nil, err
	}

	if len(response.Errors) != 0 {
		return nil, fmt.Errorf("%s", response.Errors[0].Message)
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
		client:      &http.Client{},
		bodyReader:  io.ReadAll,
		httpRequest: http.NewRequest,
		baseUrl:     settings.SonarQube.Url,
		token:       settings.SonarQube.Token.Value,
	}
}
