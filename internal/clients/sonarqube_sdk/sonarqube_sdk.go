package sonarqube_sdk

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"

	"gitea-sonarqube-pr-bot/internal/settings"
)

type SonarQubeSdkInterface interface {
	GetMeasures(string, string) (string, error)
}

type SonarQubeSdk struct {
	client  *http.Client
	baseUrl string
	token   string
}

func (sdk *SonarQubeSdk) GetMeasures(project string, branch string) (string, error) {
	url := fmt.Sprintf("%s/api/measures/component?additionalFields=metrics&metricKeys=bugs,vulnerabilities,new_security_hotspots,violations&component=%s&pullRequest=%s", sdk.baseUrl, project, branch)
	log.Println(url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(fmt.Errorf("Cannot initialize Request: %w", err))
	}
	req.Header.Add("Authorization", sdk.basicAuth())
	resp, _ := sdk.client.Do(req)

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	return string(body), nil
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
