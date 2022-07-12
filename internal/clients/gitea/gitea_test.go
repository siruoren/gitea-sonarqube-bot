package gitea

import (
	"errors"
	"net/http"
	"testing"

	"code.gitea.io/sdk/gitea"
	"codeberg.org/justusbunsi/gitea-sonarqube-bot/internal/settings"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type SdkMock struct {
	simulatedError error
	mock.Mock
}

func (m *SdkMock) CreateIssueComment(owner, repo string, index int64, opt gitea.CreateIssueCommentOption) (*gitea.Comment, *gitea.Response, error) {
	m.Called(owner, repo, index, opt)
	return nil, nil, m.simulatedError
}
func (m *SdkMock) CreateStatus(owner, repo, sha string, opts gitea.CreateStatusOption) (*gitea.Status, *gitea.Response, error) {
	m.Called(owner, repo, sha, opts)
	r := &gitea.Response{
		Response: &http.Response{
			StatusCode: http.StatusOK,
		},
	}
	if m.simulatedError != nil {
		r.StatusCode = http.StatusInternalServerError
	}
	return nil, r, m.simulatedError
}
func (m *SdkMock) GetPullRequest(owner, repo string, index int64) (*gitea.PullRequest, *gitea.Response, error) {
	m.Called(owner, repo, index)
	return &gitea.PullRequest{
		Head: &gitea.PRBranchInfo{
			Sha: "a1aada0b7b19e58ae539b4812d960bca35ev78cb",
		},
	}, nil, m.simulatedError
}

func TestNew(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		config := &settings.GiteaConfig{
			Url: "http://example.com",
			Token: &settings.Token{
				Value: "test-token",
			},
		}

		callback := func(url string, options ...gitea.ClientOption) (*SdkMock, error) {
			return &SdkMock{}, nil
		}
		assert.IsType(t, &GiteaSdk{}, New(config, callback), "")
	})

	t.Run("Initialization errors", func(t *testing.T) {
		config := &settings.GiteaConfig{
			Url: "http://example.com",
			Token: &settings.Token{
				Value: "test-token",
			},
		}

		callback := func(url string, options ...gitea.ClientOption) (*SdkMock, error) {
			return nil, errors.New("Simulated initialization error")
		}
		assert.PanicsWithError(t, "cannot initialize Gitea client: Simulated initialization error", func() { New(config, callback) })
	})
}

func TestDetermineHEAD(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		clientMock := &SdkMock{}
		clientMock.On("GetPullRequest", "test-owner", "test-repo", int64(1)).Once()

		sdk := GiteaSdk{
			client: clientMock,
		}
		sha, err := sdk.DetermineHEAD(settings.GiteaRepository{
			Owner: "test-owner",
			Name:  "test-repo",
		}, 1)

		assert.Nil(t, err)
		assert.Equal(t, "a1aada0b7b19e58ae539b4812d960bca35ev78cb", sha)
		clientMock.AssertExpectations(t)
	})

	t.Run("API error", func(t *testing.T) {
		clientMock := &SdkMock{
			simulatedError: errors.New("Simulated error"),
		}
		clientMock.On("GetPullRequest", "test-owner", "test-repo", int64(1)).Once()

		sdk := GiteaSdk{
			client: clientMock,
		}

		_, err := sdk.DetermineHEAD(settings.GiteaRepository{
			Owner: "test-owner",
			Name:  "test-repo",
		}, 1)

		assert.Errorf(t, err, "Simulated error")
		clientMock.AssertExpectations(t)
	})
}

func TestUpdateStatus(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		clientMock := &SdkMock{}
		clientMock.On("CreateStatus", "test-owner", "test-repo", "a1aada0b7b19e58ae539b4812d960bca35ev78cb", mock.Anything).Once()
		sdk := GiteaSdk{
			client: clientMock,
		}

		err := sdk.UpdateStatus(settings.GiteaRepository{
			Owner: "test-owner",
			Name:  "test-repo",
		}, "a1aada0b7b19e58ae539b4812d960bca35ev78cb", StatusDetails{
			Url:     "http://example.com",
			Message: "expected message",
			State:   StatusOK,
		})

		assert.Nil(t, err)
		clientMock.AssertExpectations(t)

		actualStatusOption := clientMock.Calls[0].Arguments[3].(gitea.CreateStatusOption)
		assert.Equal(t, "http://example.com", actualStatusOption.TargetURL)
		assert.Equal(t, "expected message", actualStatusOption.Description)
		assert.Equal(t, gitea.StatusSuccess, actualStatusOption.State)
	})

	t.Run("API error", func(t *testing.T) {
		clientMock := &SdkMock{
			simulatedError: errors.New("Simulated error"),
		}
		clientMock.On("CreateStatus", "test-owner", "test-repo", "a1aada0b7b19e58ae539b4812d960bca35ev78cb", mock.Anything).Once()
		sdk := GiteaSdk{
			client: clientMock,
		}

		err := sdk.UpdateStatus(settings.GiteaRepository{
			Owner: "test-owner",
			Name:  "test-repo",
		}, "a1aada0b7b19e58ae539b4812d960bca35ev78cb", StatusDetails{
			Url:     "http://example.com",
			Message: "expected message",
			State:   StatusOK,
		})

		assert.Errorf(t, err, "Simulated error")
		clientMock.AssertExpectations(t)
	})
}

func TestPostComment(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		clientMock := &SdkMock{}
		clientMock.On("CreateIssueComment", "test-owner", "test-repo", int64(1), mock.Anything).Once()
		sdk := GiteaSdk{
			client: clientMock,
		}

		err := sdk.PostComment(settings.GiteaRepository{
			Owner: "test-owner",
			Name:  "test-repo",
		}, 1, "test post comment")

		assert.Nil(t, err)
		clientMock.AssertExpectations(t)

		actualCommentOption := clientMock.Calls[0].Arguments[3].(gitea.CreateIssueCommentOption)
		assert.Equal(t, "test post comment", actualCommentOption.Body)
	})

	t.Run("API error", func(t *testing.T) {
		clientMock := &SdkMock{
			simulatedError: errors.New("Simulated error"),
		}
		clientMock.On("CreateIssueComment", "test-owner", "test-repo", int64(1), mock.Anything).Once()
		sdk := GiteaSdk{
			client: clientMock,
		}

		err := sdk.PostComment(settings.GiteaRepository{
			Owner: "test-owner",
			Name:  "test-repo",
		}, 1, "test post comment")

		assert.Errorf(t, err, "Simulated error")
		clientMock.AssertExpectations(t)
	})
}
