package gitea

import (
	"fmt"
	"gitea-sonarqube-pr-bot/internal/settings"
	"log"

	"code.gitea.io/sdk/gitea"
)

type GiteaSdkInterface interface {
	PostComment(settings.GiteaRepository, int, string) error
	UpdateStatus(settings.GiteaRepository, string, StatusDetails) error
	DetermineHEAD(settings.GiteaRepository, int64) (string, error)
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

func (sdk *GiteaSdk) UpdateStatus(repo settings.GiteaRepository, ref string, details StatusDetails) error {
	opt := gitea.CreateStatusOption{
		TargetURL:   details.Url,
		Context:     "gitea-sonarqube-pr-bot",
		Description: details.Message,
		State:       gitea.StatusState(details.State),
	}

	_, _, err := sdk.client.CreateStatus(repo.Owner, repo.Name, ref, opt)
	if err != nil {
		log.Printf("Error updating status: %s", err.Error())
	}

	return err
}

func (sdk *GiteaSdk) DetermineHEAD(repo settings.GiteaRepository, idx int64) (string, error) {
	pr, _, err := sdk.client.GetPullRequest(repo.Owner, repo.Name, idx)
	if err != nil {
		return "", err
	}

	return pr.Head.Sha, nil
}

func New() *GiteaSdk {
	client, err := gitea.NewClient(settings.Gitea.Url, gitea.SetToken(settings.Gitea.Token.Value))
	if err != nil {
		panic(fmt.Errorf("cannot initialize Gitea client: %w", err))
	}

	return &GiteaSdk{client}
}
