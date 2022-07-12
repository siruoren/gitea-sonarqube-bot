package gitea

import (
	"fmt"
	"log"

	"code.gitea.io/sdk/gitea"
	"codeberg.org/justusbunsi/gitea-sonarqube-bot/internal/settings"
)

type GiteaSdkInterface interface {
	PostComment(settings.GiteaRepository, int, string) error
	UpdateStatus(settings.GiteaRepository, string, StatusDetails) error
	DetermineHEAD(settings.GiteaRepository, int64) (string, error)
}

type ClientInterface interface {
	CreateIssueComment(owner, repo string, index int64, opt gitea.CreateIssueCommentOption) (*gitea.Comment, *gitea.Response, error)
	CreateStatus(owner, repo, sha string, opts gitea.CreateStatusOption) (*gitea.Status, *gitea.Response, error)
	GetPullRequest(owner, repo string, index int64) (*gitea.PullRequest, *gitea.Response, error)
}

type GiteaSdk struct {
	client ClientInterface
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
		Context:     "gitea-sonarqube-bot",
		Description: details.Message,
		State:       gitea.StatusState(details.State),
	}

	_, r, err := sdk.client.CreateStatus(repo.Owner, repo.Name, ref, opt)
	if err != nil {
		log.Printf("Error updating status: response code: %d | error: '%s'", r.StatusCode, err.Error())
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

func New[T ClientInterface](configuration *settings.GiteaConfig, newClient func(url string, options ...gitea.ClientOption) (T, error)) *GiteaSdk {
	client, err := newClient(configuration.Url, gitea.SetToken(configuration.Token.Value))
	if err != nil {
		panic(fmt.Errorf("cannot initialize Gitea client: %w", err))
	}

	return &GiteaSdk{client}
}
