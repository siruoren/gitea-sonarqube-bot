package gitea

import (
	"encoding/json"
	"fmt"
	"log"

	"codeberg.org/justusbunsi/gitea-sonarqube-bot/internal/actions"
	giteaSdk "codeberg.org/justusbunsi/gitea-sonarqube-bot/internal/clients/gitea"
	sqSdk "codeberg.org/justusbunsi/gitea-sonarqube-bot/internal/clients/sonarqube"
	"codeberg.org/justusbunsi/gitea-sonarqube-bot/internal/settings"
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

	if !actions.IsValidBotComment(w.Comment.Body) {
		return fmt.Errorf("ignore hook for non-bot action comment or unknown action")
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

	url := sqSDK.GetPullRequestUrl(w.ConfiguredProject.SonarQube.Key, w.Issue.Number)

	_ = gSDK.UpdateStatus(w.ConfiguredProject.Gitea, headRef, giteaSdk.StatusDetails{
		Url:     url,
		Message: pr.Status.QualityGateStatus,
		State:   status,
	})

	comment, err := sqSDK.ComposeGiteaComment(&sqSdk.CommentComposeData{
		Key:         w.ConfiguredProject.SonarQube.Key,
		PRName:      sqSdk.PRNameFromIndex(w.Issue.Number),
		Url:         url,
		QualityGate: pr.Status.QualityGateStatus,
	})
	if err != nil {
		return
	}
	gSDK.PostComment(w.ConfiguredProject.Gitea, int(w.Issue.Number), comment)
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
