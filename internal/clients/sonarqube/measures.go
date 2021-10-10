package sonarqube

import (
	"fmt"
	"strings"
)

type period struct {
	Value string `json:"value"`
}

type MeasuresComponentMeasure struct {
	Metric string  `json:"metric"`
	Value  string  `json:"value"`
	Period *period `json:"period,omitempty"`
}

type MeasuresComponentMetric struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type MeasuresComponent struct {
	PullRequest string                     `json:"pullRequest"`
	Measures    []MeasuresComponentMeasure `json:"measures"`
}

type MeasuresResponse struct {
	Component MeasuresComponent         `json:"component"`
	Metrics   []MeasuresComponentMetric `json:"metrics"`
}

func (mr *MeasuresResponse) GetRenderedMarkdownTable() string {
	metricsTranslations := map[string]string{}
	for _, metric := range mr.Metrics {
		metricsTranslations[metric.Key] = metric.Name
	}
	measures := make([]string, len(mr.Component.Measures))
	for i, measure := range mr.Component.Measures {
		value := measure.Value
		if measure.Period != nil {
			value = measure.Period.Value
		}
		measures[i] = fmt.Sprintf("| %s | %s |", metricsTranslations[measure.Metric], value)
	}

	table := `
| Metric | Current |
| -------- | -------- |
%s`

	return fmt.Sprintf(table, strings.Join(measures, "\n"))
}
