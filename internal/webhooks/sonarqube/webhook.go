package sonarqube

import (
	"bytes"
	"log"

	"github.com/spf13/viper"
)

type Webhook struct {
	ServerUrl string `mapstructure:"serverUrl"`
	Revision string
	Branch branch
	QualityGate qualityGate `mapstructure:"qualityGate"`
}

type branch struct {
	Name string
	Type string
	Url string
}

type qualityGate struct {
	Status string
	Conditions []condition
}

type condition struct {
	Metric string
	Status string
}

func NewWebhook(raw []byte) (*Webhook, bool) {
	v := viper.New()
	v.SetConfigType("json")
	v.ReadConfig(bytes.NewBuffer(raw))

	w := Webhook{}

	err := v.Unmarshal(&w)
	if err != nil {
	  log.Printf("Error parsing SonarQube webhook: %s", err.Error())
		return nil, false
	}

	return &w, true
}
