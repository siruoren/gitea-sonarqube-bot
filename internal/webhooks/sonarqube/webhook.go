package sonarqube

import (
	"bytes"
	"log"

	sqSdk "gitea-sonarqube-pr-bot/internal/clients/sonarqube"

	"github.com/spf13/viper"
)

type Webhook struct {
	ServerUrl string `mapstructure:"serverUrl"`
	Revision  string
	Project   struct {
		Key  string
		Name string
		Url  string
	}
	Branch struct {
		Name string
		Type string
		Url  string
	}
	QualityGate struct {
		Status     string
		Conditions []struct {
			Metric string
			Status string
		}
	} `mapstructure:"qualityGate"`
	PRIndex int
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

	idx, err1 := sqSdk.ParsePRIndex(w.Branch.Name)
	if err1 != nil {
		log.Printf("Error parsing PR index: %s", err1.Error())
		return w, false
	}

	w.PRIndex = idx

	return w, true
}
