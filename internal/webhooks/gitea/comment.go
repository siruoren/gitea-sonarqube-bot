package gitea

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"gitea-sonarqube-pr-bot/internal/actions"
	giteaSdk "gitea-sonarqube-pr-bot/internal/clients/gitea"
	sqSdk "gitea-sonarqube-pr-bot/internal/clients/sonarqube"
	"gitea-sonarqube-pr-bot/internal/settings"
)

type issue struct {
	Number     int64                    `json:"number"`
	Repository settings.GiteaRepository `json:"repository"`
}

type comment struct {
	Body string `json:"body"`
}

type CommentWebhook struct {
	Action            string  `json:"action"`
	IsPR              bool    `json:"is_pull"`
	Issue             issue   `json:"issue"`
	Comment           comment `json:"comment"`
	ConfiguredProject settings.Project
}

func (w *CommentWebhook) inProjectsMapping(p []settings.Project) (bool, int) {
	owner := w.Issue.Repository.Owner
	name := w.Issue.Repository.Name
	for idx, proj := range p {
		if proj.Gitea.Owner == owner && proj.Gitea.Name == name {
			return true, idx
		}
	}

	return false, 0
}

func (w *CommentWebhook) Validate() error {
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

	if !strings.HasPrefix(w.Comment.Body, actions.ActionPrefix) {
		return fmt.Errorf("ignore hook for non-bot action comment")
	}

	w.ConfiguredProject = settings.Projects[pIdx]

	return nil
}

func (w *CommentWebhook) ProcessData(gSDK giteaSdk.GiteaSdkInterface, sqSDK sqSdk.SonarQubeSdkInterface) {
	headRef, err := gSDK.DetermineHEAD(w.ConfiguredProject.Gitea, w.Issue.Number)
	if err != nil {
		log.Printf("Error retrieving HEAD ref: %s", err.Error())
		return
	}
	log.Printf("Fetching SonarQube data...")

	pr, err := sqSDK.GetPullRequest(w.ConfiguredProject.SonarQube.Key, w.Issue.Number)
	if err != nil {
		log.Printf("Error loading PR data from SonarQube: %s", err.Error())
		return
	}

	status := giteaSdk.StatusOK
	if pr.Status.QualityGateStatus != "OK" {
		status = giteaSdk.StatusFailure
	}

	_ = gSDK.UpdateStatus(w.ConfiguredProject.Gitea, headRef, giteaSdk.StatusDetails{
		Url:     sqSDK.GetPullRequestUrl(w.ConfiguredProject.SonarQube.Key, w.Issue.Number),
		Message: pr.Status.QualityGateStatus,
		State:   status,
	})
}

func NewCommentWebhook(raw []byte) (*CommentWebhook, bool) {
	w := &CommentWebhook{}
	err := json.Unmarshal(raw, &w)
	if err != nil {
		log.Printf("Error parsing Gitea webhook: %s", err.Error())
		return w, false
	}

	return w, true
}
