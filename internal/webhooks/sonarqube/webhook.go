package sonarqube

import (
	"bytes"
	"log"

	"github.com/spf13/viper"
)

type Webhook struct {
	ServerUrl string `mapstructure:"serverUrl"`
	Revision string
	Branch struct {
		Name string
		Type string
		Url string
  }
	QualityGate struct {
		Status string
		Conditions []struct {
			Metric string
			Status string
		}
	} `mapstructure:"qualityGate"`
}

func New(raw []byte) (*Webhook, bool) {
	v := viper.New()
	v.SetConfigType("json")
	v.ReadConfig(bytes.NewBuffer(raw))

	w := &Webhook{}

	err := v.Unmarshal(&w)
	if err != nil {
	  log.Printf("Error parsing SonarQube webhook: %s", err.Error())
		return w, false
	}

	return w, true
}
