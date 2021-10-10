package gitea

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	giteaSdk "gitea-sonarqube-pr-bot/internal/clients/gitea_sdk"
	sqSdk "gitea-sonarqube-pr-bot/internal/clients/sonarqube_sdk"
	"gitea-sonarqube-pr-bot/internal/settings"
)

type BotAction string

const (
	ActionReview BotAction = "/pr-bot review"
)

type issue struct {
	Number     int64                    `json:"number"`
	Repository settings.GiteaRepository `json:"repository"`
}

type comment struct {
	Body string `json:"body"`
}

type Webhook struct {
	Action            string  `json:"action"`
	IsPR              bool    `json:"is_pull"`
	Issue             issue   `json:"issue"`
	Comment           comment `json:"comment"`
	ConfiguredProject settings.Project
}

func (w *Webhook) inProjectsMapping(p []settings.Project) (bool, int) {
	owner := w.Issue.Repository.Owner
	name := w.Issue.Repository.Name
	for idx, proj := range p {
		if proj.Gitea.Owner == owner && proj.Gitea.Name == name {
			return true, idx
		}
	}

	return false, 0
}

func (w *Webhook) Validate() error {
	if !w.IsPR {
		return fmt.Errorf("ignore non-PR hook")
	}

	found, pIdx := w.inProjectsMapping(settings.Projects)
	if !found {
		return fmt.Errorf("ignore hook for non-configured project '%s/%s'", w.Issue.Repository.Owner, w.Issue.Repository.Name)
	}

	if w.Action != "created" {
		return fmt.Errorf("ignore hook for action others than created")
	}

	if !strings.HasPrefix(w.Comment.Body, "/pr-bot") {
		return fmt.Errorf("ignore hook for non-bot action comment")
	}

	w.ConfiguredProject = settings.Projects[pIdx]

	return nil
}

func (w *Webhook) ProcessData(gSDK giteaSdk.GiteaSdkInterface, sqSDK sqSdk.SonarQubeSdkInterface) {
	headRef, err := gSDK.DetermineHEAD(w.ConfiguredProject.Gitea, w.Issue.Number)
	if err != nil {
		log.Printf("Error retrieving HEAD ref: %s", err.Error())
		return
	}
	log.Printf("Fetching SonarQube data...")

	_ = gSDK.UpdateStatus(w.ConfiguredProject.Gitea, headRef, giteaSdk.StatusDetails{
		Url:     "",
		Message: "OK",
		State:   giteaSdk.StatusOK,
	})
}

func New(raw []byte) (*Webhook, bool) {
	w := &Webhook{}
	err := json.Unmarshal(raw, &w)
	if err != nil {
		log.Printf("Error parsing Gitea webhook: %s", err.Error())
		return w, false
	}

	return w, true
}
