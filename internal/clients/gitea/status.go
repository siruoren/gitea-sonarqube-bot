package gitea

import "code.gitea.io/sdk/gitea"

type State gitea.StatusState

const (
	StatusOK      State = State(gitea.StatusSuccess)
	StatusPending State = State(gitea.StatusPending)
	StatusFailure State = State(gitea.StatusFailure)
)

type StatusDetails struct {
	Url     string
	Message string
	State   State
}
