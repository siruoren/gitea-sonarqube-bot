package gitea_sdk

import "code.gitea.io/sdk/gitea"

type State gitea.StatusState

const (
	StatusOK      State = State(gitea.StatusSuccess)
	StatusFailure State = State(gitea.StatusFailure)
)

type StatusDetails struct {
	Url     string
	Message string
	State   State
}
