package settings

import "strings"

type sonarQubeConfig struct {
	Url               string
	Token             *Token
	Webhook           *Webhook
	AdditionalMetrics []string
}

func (c *sonarQubeConfig) GetMetricsList() string {
	metrics := []string{
		"bugs",
		"vulnerabilities",
		"code_smells",
	}
	if len(c.AdditionalMetrics) != 0 {
		metrics = append(metrics, c.AdditionalMetrics...)
	}
	return strings.Join(metrics, ",")
}
