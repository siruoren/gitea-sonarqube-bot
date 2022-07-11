package sonarqube

import (
	"encoding/json"
	"log"

	sqSdk "gitea-sonarqube-bot/internal/clients/sonarqube"
)

type properties struct {
	OriginalCommit string `json:"sonar.analysis.sqbot,omitempty"`
}

type Webhook struct {
	ServerUrl string `json:"serverUrl"`
	Revision  string `json:"revision"`
	Project   struct {
		Key  string `json:"key"`
		Name string `json:"name"`
		Url  string `json:"url"`
	} `json:"project"`
	Branch struct {
		Name string `json:"name"`
		Type string `json:"type"`
		Url  string `json:"url"`
	} `json:"branch"`
	QualityGate struct {
		Status     string `json:"status"`
		Conditions []struct {
			Metric string
			Status string
		} `json:"conditions"`
	} `json:"qualityGate"`
	Properties *properties `json:"properties,omitempty"`
	PRIndex    int
}

func (w *Webhook) GetRevision() string {
	if w.Properties != nil && w.Properties.OriginalCommit != "" {
		return w.Properties.OriginalCommit
	}

	return w.Revision
}

func New(raw []byte) (*Webhook, bool) {
	w := &Webhook{}

	err := json.Unmarshal(raw, w)
	if err != nil {
		log.Printf("Error parsing SonarQube webhook: %s", err.Error())
		return w, false
	}

	idx, err1 := sqSdk.ParsePRIndex(w.Branch.Name)
	if err1 != nil {
		log.Printf("Error parsing PR index: %s", err1.Error())
		return w, false
	}

	w.PRIndex = idx

	return w, true
}
