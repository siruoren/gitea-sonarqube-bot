package gitea_sdk

import (
	"fmt"
	"gitea-sonarqube-pr-bot/internal/settings"
	webhook "gitea-sonarqube-pr-bot/internal/webhooks/sonarqube"

	"code.gitea.io/sdk/gitea"
)

type GiteaSdkInterface interface {
	PostComment(settings.GiteaRepository, int, string) error
	UpdateStatus(settings.GiteaRepository, *webhook.Webhook) error
}

type GiteaSdk struct {
	client *gitea.Client
}

func (sdk *GiteaSdk) PostComment(repo settings.GiteaRepository, idx int, msg string) error {
	opt := gitea.CreateIssueCommentOption{
		Body: msg,
	}

	_, _, err := sdk.client.CreateIssueComment(repo.Owner, repo.Name, int64(idx), opt)

	return err
}

func (sdk *GiteaSdk) UpdateStatus(repo settings.GiteaRepository, w *webhook.Webhook) error {
	status := gitea.StatusPending
	switch w.QualityGate.Status {
	case "OK":
		status = gitea.StatusSuccess
	case "ERROR":
		status = gitea.StatusFailure
	}
	opt := gitea.CreateStatusOption{
		TargetURL:   w.Branch.Url,
		Context:     "gitea-sonarqube-pr-bot",
		Description: w.QualityGate.Status,
		State:       status,
	}

	_, _, err := sdk.client.CreateStatus(repo.Owner, repo.Name, w.Revision, opt)

	return err
}

func New() *GiteaSdk {
	client, err := gitea.NewClient(settings.Gitea.Url, gitea.SetToken(settings.Gitea.Token.Value))
	if err != nil {
		panic(fmt.Errorf("cannot initialize Gitea client: %w", err))
	}

	return &GiteaSdk{client}
}
