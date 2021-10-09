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

func (w *Webhook) GetRenderedQualityGate() string {
	status := ":white_check_mark:"
	if w.QualityGate.Status != "OK" {
		status = ":x:"
	}

	return fmt.Sprintf("**Quality Gate**: %s", status)
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

	idx, err1 := parsePRIndex(w)
	if err1 != nil {
		log.Printf("Error parsing PR index: %s", err1.Error())
		return w, false
	}

	w.PRIndex = idx

	return w, true
}

func parsePRIndex(w *Webhook) (int, error) {
	re := regexp.MustCompile(`^PR-(\d+)$`)
	res := re.FindSubmatch([]byte(w.Branch.Name))
	if len(res) != 2 {
		return 0, fmt.Errorf("branch name '%s' does not match regex '%s'", w.Branch.Name, re.String())
	}

	return strconv.Atoi(string(res[1]))
}
