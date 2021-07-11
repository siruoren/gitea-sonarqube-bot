package sonarqube

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
	"strconv"

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

func (w *Webhook) GetPRIndex() (int, error) {
	re := regexp.MustCompile(`^PR-(\d+)$`)
	res := re.FindSubmatch([]byte(w.Branch.Name))
	if len(res) != 2 {
		return 0, fmt.Errorf("Branch name '%s' does not match regex '%s'. Extracting PR index not possible.", w.Branch.Name, re.String())
	}

	return strconv.Atoi(fmt.Sprintf("%s", res[1]))
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
