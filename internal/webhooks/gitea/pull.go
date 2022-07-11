package gitea

import (
	"encoding/json"
	"fmt"
	"log"

	giteaSdk "gitea-sonarqube-bot/internal/clients/gitea"
	sqSdk "gitea-sonarqube-bot/internal/clients/sonarqube"
	"gitea-sonarqube-bot/internal/settings"
)

type pullRequest struct {
	Number int64 `json:"number"`
	Head   struct {
		Sha string `json:"sha"`
	} `json:"head"`
}

type repoOwner struct {
	Login string `json:"login"`
}

type rawRepository struct {
	Name  string    `json:"name"`
	Owner repoOwner `json:"owner"`
}

type PullWebhook struct {
	Action            string        `json:"action"`
	PullRequest       pullRequest   `json:"pull_request"`
	RawRepository     rawRepository `json:"repository"`
	Repository        settings.GiteaRepository
	ConfiguredProject settings.Project
}

func (w *PullWebhook) inProjectsMapping(p []settings.Project) (bool, int) {
	owner := w.RawRepository.Owner.Login
	name := w.RawRepository.Name
	for idx, proj := range p {
		if proj.Gitea.Owner == owner && proj.Gitea.Name == name {
			return true, idx
		}
	}

	return false, 0
}

func (w *PullWebhook) Validate() error {
	found, pIdx := w.inProjectsMapping(settings.Projects)
	owner := w.RawRepository.Owner.Login
	name := w.RawRepository.Name
	if !found {
		return fmt.Errorf("ignore hook for non-configured project '%s/%s'", owner, name)
	}

	if w.Action != "synchronized" && w.Action != "opened" {
		return fmt.Errorf("ignore hook for action others than 'opened' or 'synchronized'")
	}

	w.Repository = settings.GiteaRepository{
		Owner: owner,
		Name:  name,
	}
	w.ConfiguredProject = settings.Projects[pIdx]

	return nil
}

func (w *PullWebhook) ProcessData(gSDK giteaSdk.GiteaSdkInterface, sqSDK sqSdk.SonarQubeSdkInterface) {
	_ = gSDK.UpdateStatus(w.ConfiguredProject.Gitea, w.PullRequest.Head.Sha, giteaSdk.StatusDetails{
		Url:     "",
		Message: "Analysis pending...",
		State:   giteaSdk.StatusPending,
	})
}

func NewPullWebhook(raw []byte) (*PullWebhook, bool) {
	w := &PullWebhook{}
	err := json.Unmarshal(raw, &w)
	if err != nil {
		log.Printf("Error parsing Gitea webhook: %s", err.Error())
		return w, false
	}

	return w, true
}
