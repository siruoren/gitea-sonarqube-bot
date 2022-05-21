package settings

import "strings"

type SonarQubeConfig struct {
	Url               string
	Token             *Token
	Webhook           *Webhook
	AdditionalMetrics []string
}

func (c *SonarQubeConfig) GetMetricsList() string {
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
