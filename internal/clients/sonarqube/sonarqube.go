package sonarqube

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"gitea-sonarqube-pr-bot/internal/settings"
)

type SonarQubeSdkInterface interface {
	GetMeasures(string, string) (*MeasuresResponse, error)
}

type SonarQubeSdk struct {
	client  *http.Client
	baseUrl string
	token   string
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
