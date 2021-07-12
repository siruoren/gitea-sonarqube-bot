package gitea_sdk

import (
	"fmt"
	"log"
	"gitea-sonarqube-pr-bot/internal/settings"
	"code.gitea.io/sdk/gitea"
)

type GiteaSdkInterface interface {
	PostComment(settings.GiteaRepository, int, string) error
}

type GiteaSdk struct {
	client *gitea.Client
}

func (sdk *GiteaSdk) PostComment(repo settings.GiteaRepository, idx int, msg string) error {
	opt := gitea.CreateIssueCommentOption{
		Body: msg,
	}

	log.Printf("Owner: %s | Repo: %s | Index: %d | Options: %s", repo.Owner, repo.Name, idx, opt)
	_, _, err := sdk.client.CreateIssueComment(repo.Owner, repo.Name, int64(idx), opt)

	return err
}

func New() *GiteaSdk {
	client, err := gitea.NewClient(settings.Gitea.Url, gitea.SetToken(settings.Gitea.Token.Value))
	if err != nil {
		panic(fmt.Errorf("Cannot initialize Gitea client: %w", err))
	}

	return &GiteaSdk{client}
}
