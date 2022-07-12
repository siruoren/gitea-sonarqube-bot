package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"codeberg.org/justusbunsi/gitea-sonarqube-bot/internal/settings"
	"github.com/stretchr/testify/assert"
)

func TestHandleGiteaCommentWebhook(t *testing.T) {
	withValidRequestData := func(t *testing.T, jsonBody []byte) (*http.Request, *httptest.ResponseRecorder, http.HandlerFunc) {
		webhookHandler := NewGiteaWebhookHandler(new(GiteaSdkMock), new(SQSdkMock))

		req, err := http.NewRequest("POST", "/hooks/gitea", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			status, response := webhookHandler.HandleComment(r)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			io.WriteString(w, fmt.Sprintf(`{"message": "%s"}`, response))
		})

		return req, rr, handler
	}

	t.Run("On success", func(t *testing.T) {
		settings.Pattern = &settings.PatternConfig{
			Template: "PR-%d",
		}
		settings.Gitea = settings.GiteaConfig{
			Webhook: &settings.Webhook{
				Secret: "",
			},
		}
		settings.Projects = []settings.Project{
			{
				SonarQube: struct{ Key string }{
					Key: "gitea-sonarqube-bot",
				},
				Gitea: settings.GiteaRepository{
					Owner: "test-user",
					Name:  "gitea-sonarqube-bot",
				},
			},
		}
		req, rr, handler := withValidRequestData(t, []byte(`{"action":"created","issue":{"id":1,"url":"http://localhost:3000/api/v1/repos/test-user/gitea-sonarqube-bot/issues/1","html_url":"http://localhost:3000/test-user/gitea-sonarqube-bot/pulls/1","number":1,"user":{"id":1,"login":"test-user","full_name":"","email":"a@b.c","avatar_url":"http://localhost:3000/avatar/5d60d4e28066df254d5452f92c910092","language":"","is_admin":false,"last_login":"0001-01-01T00:00:00Z","created":"2022-05-15T18:42:54Z","restricted":false,"active":false,"prohibit_login":false,"location":"","website":"","description":"","visibility":"public","followers_count":0,"following_count":0,"starred_repos_count":0,"username":"test-user"},"original_author":"","original_author_id":0,"title":"„README.md“ ändern","body":"","ref":"","labels":[],"milestone":null,"assignee":null,"assignees":null,"state":"open","is_locked":false,"comments":0,"created_at":"2022-05-15T18:46:19Z","updated_at":"2022-05-15T18:57:29Z","closed_at":null,"due_date":null,"pull_request":{"merged":false,"merged_at":null},"repository":{"id":1,"name":"gitea-sonarqube-bot","owner":"test-user","full_name":"test-user/gitea-sonarqube-bot"}},"comment":{"id":2,"html_url":"http://localhost:3000/test-user/gitea-sonarqube-bot/pulls/1#issuecomment-2","pull_request_url":"http://localhost:3000/test-user/gitea-sonarqube-bot/pulls/1","issue_url":"","user":{"id":1,"login":"test-user","full_name":"","email":"a@b.c","avatar_url":"http://localhost:3000/avatar/5d60d4e28066df254d5452f92c910092","language":"","is_admin":false,"last_login":"0001-01-01T00:00:00Z","created":"2022-05-15T18:42:54Z","restricted":false,"active":false,"prohibit_login":false,"location":"","website":"","description":"","visibility":"public","followers_count":0,"following_count":0,"starred_repos_count":0,"username":"test-user"},"original_author":"","original_author_id":0,"body":"/sq-bot review","created_at":"2022-05-15T18:57:29Z","updated_at":"2022-05-15T18:57:29Z"},"repository":{"id":1,"owner":{"id":1,"login":"test-user","full_name":"","email":"a@b.c","avatar_url":"http://localhost:3000/avatar/5d60d4e28066df254d5452f92c910092","language":"","is_admin":false,"last_login":"0001-01-01T00:00:00Z","created":"2022-05-15T18:42:54Z","restricted":false,"active":false,"prohibit_login":false,"location":"","website":"","description":"","visibility":"public","followers_count":0,"following_count":0,"starred_repos_count":0,"username":"test-user"},"name":"gitea-sonarqube-bot","full_name":"test-user/gitea-sonarqube-bot","description":"","empty":false,"private":false,"fork":false,"template":false,"parent":null,"mirror":false,"size":110,"html_url":"http://localhost:3000/test-user/gitea-sonarqube-bot","ssh_url":"git@localhost:test-user/gitea-sonarqube-bot.git","clone_url":"http://localhost:3000/test-user/gitea-sonarqube-bot.git","original_url":"","website":"","stars_count":0,"forks_count":0,"watchers_count":1,"open_issues_count":0,"open_pr_counter":1,"release_counter":0,"default_branch":"main","archived":false,"created_at":"2022-05-15T18:45:46Z","updated_at":"2022-05-15T18:46:09Z","permissions":{"admin":true,"push":true,"pull":true},"has_issues":true,"internal_tracker":{"enable_time_tracker":true,"allow_only_contributors_to_track_time":true,"enable_issue_dependencies":true},"has_wiki":true,"has_pull_requests":true,"has_projects":true,"ignore_whitespace_conflicts":false,"allow_merge_commits":true,"allow_rebase":true,"allow_rebase_explicit":true,"allow_squash_merge":true,"default_merge_style":"merge","avatar_url":"","internal":false,"mirror_interval":"","mirror_updated":"0001-01-01T00:00:00Z","repo_transfer":null},"sender":{"id":1,"login":"test-user","full_name":"","email":"a@b.c","avatar_url":"http://localhost:3000/avatar/5d60d4e28066df254d5452f92c910092","language":"","is_admin":false,"last_login":"0001-01-01T00:00:00Z","created":"2022-05-15T18:42:54Z","restricted":false,"active":false,"prohibit_login":false,"location":"","website":"","description":"","visibility":"public","followers_count":0,"following_count":0,"starred_repos_count":0,"username":"test-user"},"is_pull":true}`))
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, `{"message": "Processing data. See bot logs for details."}`, rr.Body.String())

		t.Cleanup(func() {
			settings.Pattern = nil
		})
	})

	t.Run("With invalid JSON body", func(t *testing.T) {
		settings.Pattern = &settings.PatternConfig{
			Template: "PR-%d",
		}
		settings.Gitea = settings.GiteaConfig{
			Webhook: &settings.Webhook{
				Secret: "",
			},
		}
		settings.Projects = []settings.Project{
			{
				SonarQube: struct{ Key string }{
					Key: "gitea-sonarqube-bot",
				},
			},
		}

		req, rr, handler := withValidRequestData(t, []byte(`{ "action": ["non-string-value-for-action"] }`))
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
		assert.Equal(t, `{"message": "Error parsing POST body."}`, rr.Body.String())

		t.Cleanup(func() {
			settings.Pattern = nil
		})
	})

	t.Run("With invalid signature", func(t *testing.T) {
		settings.Gitea = settings.GiteaConfig{
			Webhook: &settings.Webhook{
				Secret: "gitea-comment-test-webhook",
			},
		}
		settings.Projects = []settings.Project{
			{
				SonarQube: struct{ Key string }{
					Key: "pr-bot",
				},
			},
		}
		req, rr, handler := withValidRequestData(t, []byte(`{"action":"created","issue":{"id":1,"url":"http://localhost:3000/api/v1/repos/test-user/gitea-sonarqube-bot/issues/1","html_url":"http://localhost:3000/test-user/gitea-sonarqube-bot/pulls/1","number":1,"user":{"id":1,"login":"test-user","full_name":"","email":"a@b.c","avatar_url":"http://localhost:3000/avatar/5d60d4e28066df254d5452f92c910092","language":"","is_admin":false,"last_login":"0001-01-01T00:00:00Z","created":"2022-05-15T18:42:54Z","restricted":false,"active":false,"prohibit_login":false,"location":"","website":"","description":"","visibility":"public","followers_count":0,"following_count":0,"starred_repos_count":0,"username":"test-user"},"original_author":"","original_author_id":0,"title":"„README.md“ ändern","body":"","ref":"","labels":[],"milestone":null,"assignee":null,"assignees":null,"state":"open","is_locked":false,"comments":0,"created_at":"2022-05-15T18:46:19Z","updated_at":"2022-05-15T18:57:29Z","closed_at":null,"due_date":null,"pull_request":{"merged":false,"merged_at":null},"repository":{"id":1,"name":"gitea-sonarqube-bot","owner":"test-user","full_name":"test-user/gitea-sonarqube-bot"}},"comment":{"id":2,"html_url":"http://localhost:3000/test-user/gitea-sonarqube-bot/pulls/1#issuecomment-2","pull_request_url":"http://localhost:3000/test-user/gitea-sonarqube-bot/pulls/1","issue_url":"","user":{"id":1,"login":"test-user","full_name":"","email":"a@b.c","avatar_url":"http://localhost:3000/avatar/5d60d4e28066df254d5452f92c910092","language":"","is_admin":false,"last_login":"0001-01-01T00:00:00Z","created":"2022-05-15T18:42:54Z","restricted":false,"active":false,"prohibit_login":false,"location":"","website":"","description":"","visibility":"public","followers_count":0,"following_count":0,"starred_repos_count":0,"username":"test-user"},"original_author":"","original_author_id":0,"body":"/sq-bot review","created_at":"2022-05-15T18:57:29Z","updated_at":"2022-05-15T18:57:29Z"},"repository":{"id":1,"owner":{"id":1,"login":"test-user","full_name":"","email":"a@b.c","avatar_url":"http://localhost:3000/avatar/5d60d4e28066df254d5452f92c910092","language":"","is_admin":false,"last_login":"0001-01-01T00:00:00Z","created":"2022-05-15T18:42:54Z","restricted":false,"active":false,"prohibit_login":false,"location":"","website":"","description":"","visibility":"public","followers_count":0,"following_count":0,"starred_repos_count":0,"username":"test-user"},"name":"gitea-sonarqube-bot","full_name":"test-user/gitea-sonarqube-bot","description":"","empty":false,"private":false,"fork":false,"template":false,"parent":null,"mirror":false,"size":110,"html_url":"http://localhost:3000/test-user/gitea-sonarqube-bot","ssh_url":"git@localhost:test-user/gitea-sonarqube-bot.git","clone_url":"http://localhost:3000/test-user/gitea-sonarqube-bot.git","original_url":"","website":"","stars_count":0,"forks_count":0,"watchers_count":1,"open_issues_count":0,"open_pr_counter":1,"release_counter":0,"default_branch":"main","archived":false,"created_at":"2022-05-15T18:45:46Z","updated_at":"2022-05-15T18:46:09Z","permissions":{"admin":true,"push":true,"pull":true},"has_issues":true,"internal_tracker":{"enable_time_tracker":true,"allow_only_contributors_to_track_time":true,"enable_issue_dependencies":true},"has_wiki":true,"has_pull_requests":true,"has_projects":true,"ignore_whitespace_conflicts":false,"allow_merge_commits":true,"allow_rebase":true,"allow_rebase_explicit":true,"allow_squash_merge":true,"default_merge_style":"merge","avatar_url":"","internal":false,"mirror_interval":"","mirror_updated":"0001-01-01T00:00:00Z","repo_transfer":null},"sender":{"id":1,"login":"test-user","full_name":"","email":"a@b.c","avatar_url":"http://localhost:3000/avatar/5d60d4e28066df254d5452f92c910092","language":"","is_admin":false,"last_login":"0001-01-01T00:00:00Z","created":"2022-05-15T18:42:54Z","restricted":false,"active":false,"prohibit_login":false,"location":"","website":"","description":"","visibility":"public","followers_count":0,"following_count":0,"starred_repos_count":0,"username":"test-user"},"is_pull":true}`))
		req.Header.Set("X-Gitea-Signature", "647f2395d30b1b7efcb58d9338be5b69c2addb54faf6bde6314a57ea28f45467")
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusPreconditionFailed, rr.Code)
		assert.Equal(t, `{"message": "Webhook validation failed. Request rejected."}`, rr.Body.String())
	})

	t.Run("With ignored project", func(t *testing.T) {
		settings.Gitea = settings.GiteaConfig{
			Webhook: &settings.Webhook{
				Secret: "",
			},
		}
		settings.Projects = []settings.Project{
			{
				SonarQube: struct{ Key string }{
					Key: "gitea-sonarqube-bot",
				},
			},
		}
		req, rr, handler := withValidRequestData(t, []byte(`{"action":"created","issue":{"id":1,"url":"http://localhost:3000/api/v1/repos/test-user/gitea-sonarqube-bot/issues/1","html_url":"http://localhost:3000/test-user/gitea-sonarqube-bot/pulls/1","number":1,"user":{"id":1,"login":"test-user","full_name":"","email":"a@b.c","avatar_url":"http://localhost:3000/avatar/5d60d4e28066df254d5452f92c910092","language":"","is_admin":false,"last_login":"0001-01-01T00:00:00Z","created":"2022-05-15T18:42:54Z","restricted":false,"active":false,"prohibit_login":false,"location":"","website":"","description":"","visibility":"public","followers_count":0,"following_count":0,"starred_repos_count":0,"username":"test-user"},"original_author":"","original_author_id":0,"title":"„README.md“ ändern","body":"","ref":"","labels":[],"milestone":null,"assignee":null,"assignees":null,"state":"open","is_locked":false,"comments":0,"created_at":"2022-05-15T18:46:19Z","updated_at":"2022-05-15T18:57:29Z","closed_at":null,"due_date":null,"pull_request":{"merged":false,"merged_at":null},"repository":{"id":1,"name":"gitea-sonarqube-bot","owner":"test-user","full_name":"test-user/gitea-sonarqube-bot"}},"comment":{"id":2,"html_url":"http://localhost:3000/test-user/gitea-sonarqube-bot/pulls/1#issuecomment-2","pull_request_url":"http://localhost:3000/test-user/gitea-sonarqube-bot/pulls/1","issue_url":"","user":{"id":1,"login":"test-user","full_name":"","email":"a@b.c","avatar_url":"http://localhost:3000/avatar/5d60d4e28066df254d5452f92c910092","language":"","is_admin":false,"last_login":"0001-01-01T00:00:00Z","created":"2022-05-15T18:42:54Z","restricted":false,"active":false,"prohibit_login":false,"location":"","website":"","description":"","visibility":"public","followers_count":0,"following_count":0,"starred_repos_count":0,"username":"test-user"},"original_author":"","original_author_id":0,"body":"/sq-bot review","created_at":"2022-05-15T18:57:29Z","updated_at":"2022-05-15T18:57:29Z"},"repository":{"id":1,"owner":{"id":1,"login":"test-user","full_name":"","email":"a@b.c","avatar_url":"http://localhost:3000/avatar/5d60d4e28066df254d5452f92c910092","language":"","is_admin":false,"last_login":"0001-01-01T00:00:00Z","created":"2022-05-15T18:42:54Z","restricted":false,"active":false,"prohibit_login":false,"location":"","website":"","description":"","visibility":"public","followers_count":0,"following_count":0,"starred_repos_count":0,"username":"test-user"},"name":"gitea-sonarqube-bot","full_name":"test-user/gitea-sonarqube-bot","description":"","empty":false,"private":false,"fork":false,"template":false,"parent":null,"mirror":false,"size":110,"html_url":"http://localhost:3000/test-user/gitea-sonarqube-bot","ssh_url":"git@localhost:test-user/gitea-sonarqube-bot.git","clone_url":"http://localhost:3000/test-user/gitea-sonarqube-bot.git","original_url":"","website":"","stars_count":0,"forks_count":0,"watchers_count":1,"open_issues_count":0,"open_pr_counter":1,"release_counter":0,"default_branch":"main","archived":false,"created_at":"2022-05-15T18:45:46Z","updated_at":"2022-05-15T18:46:09Z","permissions":{"admin":true,"push":true,"pull":true},"has_issues":true,"internal_tracker":{"enable_time_tracker":true,"allow_only_contributors_to_track_time":true,"enable_issue_dependencies":true},"has_wiki":true,"has_pull_requests":true,"has_projects":true,"ignore_whitespace_conflicts":false,"allow_merge_commits":true,"allow_rebase":true,"allow_rebase_explicit":true,"allow_squash_merge":true,"default_merge_style":"merge","avatar_url":"","internal":false,"mirror_interval":"","mirror_updated":"0001-01-01T00:00:00Z","repo_transfer":null},"sender":{"id":1,"login":"test-user","full_name":"","email":"a@b.c","avatar_url":"http://localhost:3000/avatar/5d60d4e28066df254d5452f92c910092","language":"","is_admin":false,"last_login":"0001-01-01T00:00:00Z","created":"2022-05-15T18:42:54Z","restricted":false,"active":false,"prohibit_login":false,"location":"","website":"","description":"","visibility":"public","followers_count":0,"following_count":0,"starred_repos_count":0,"username":"test-user"},"is_pull":true}`))
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, `{"message": "ignore hook for non-configured project 'test-user/gitea-sonarqube-bot'"}`, rr.Body.String())
	})
}

func TestHandleGiteaSynchronizeWebhook(t *testing.T) {
	withValidRequestData := func(t *testing.T, jsonBody []byte) (*http.Request, *httptest.ResponseRecorder, http.HandlerFunc) {
		webhookHandler := NewGiteaWebhookHandler(new(GiteaSdkMock), new(SQSdkMock))

		req, err := http.NewRequest("POST", "/hooks/gitea", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			status, response := webhookHandler.HandleSynchronize(r)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			io.WriteString(w, fmt.Sprintf(`{"message": "%s"}`, response))
		})

		return req, rr, handler
	}

	t.Run("On success", func(t *testing.T) {
		settings.Gitea = settings.GiteaConfig{
			Webhook: &settings.Webhook{
				Secret: "",
			},
		}
		settings.Projects = []settings.Project{
			{
				SonarQube: struct{ Key string }{
					Key: "gitea-sonarqube-bot",
				},
				Gitea: settings.GiteaRepository{
					Owner: "test-user",
					Name:  "gitea-sonarqube-bot",
				},
			},
		}
		req, rr, handler := withValidRequestData(t, []byte(`{"action":"opened","number":1,"pull_request":{"id":1,"url":"http://localhost:3000/test-user/gitea-sonarqube-bot/pulls/1","number":1,"user":{"id":1,"login":"test-user","full_name":"","email":"a@b.c","avatar_url":"http://localhost:3000/avatar/5d60d4e28066df254d5452f92c910092","language":"","is_admin":false,"last_login":"0001-01-01T00:00:00Z","created":"2022-05-15T18:42:54Z","restricted":false,"active":false,"prohibit_login":false,"location":"","website":"","description":"","visibility":"public","followers_count":0,"following_count":0,"starred_repos_count":0,"username":"test-user"},"title":"„README.md“ ändern","body":"","labels":[],"milestone":null,"assignee":null,"assignees":null,"state":"open","is_locked":false,"comments":0,"html_url":"http://localhost:3000/test-user/gitea-sonarqube-bot/pulls/1","diff_url":"http://localhost:3000/test-user/gitea-sonarqube-bot/pulls/1.diff","patch_url":"http://localhost:3000/test-user/gitea-sonarqube-bot/pulls/1.patch","mergeable":true,"merged":false,"merged_at":null,"merge_commit_sha":null,"merged_by":null,"base":{"label":"main","ref":"main","sha":"2e5c9f7fe85fd8fb6019b3dd299744e0afce076b","repo_id":1,"repo":{"id":1,"owner":{"id":1,"login":"test-user","full_name":"","email":"a@b.c","avatar_url":"http://localhost:3000/avatar/5d60d4e28066df254d5452f92c910092","language":"","is_admin":false,"last_login":"0001-01-01T00:00:00Z","created":"2022-05-15T18:42:54Z","restricted":false,"active":false,"prohibit_login":false,"location":"","website":"","description":"","visibility":"public","followers_count":0,"following_count":0,"starred_repos_count":0,"username":"test-user"},"name":"gitea-sonarqube-bot","full_name":"test-user/gitea-sonarqube-bot","description":"","empty":false,"private":false,"fork":false,"template":false,"parent":null,"mirror":false,"size":110,"html_url":"http://localhost:3000/test-user/gitea-sonarqube-bot","ssh_url":"git@localhost:test-user/gitea-sonarqube-bot.git","clone_url":"http://localhost:3000/test-user/gitea-sonarqube-bot.git","original_url":"","website":"","stars_count":0,"forks_count":0,"watchers_count":1,"open_issues_count":0,"open_pr_counter":0,"release_counter":0,"default_branch":"main","archived":false,"created_at":"2022-05-15T18:45:46Z","updated_at":"2022-05-15T18:46:09Z","permissions":{"admin":false,"push":false,"pull":true},"has_issues":true,"internal_tracker":{"enable_time_tracker":true,"allow_only_contributors_to_track_time":true,"enable_issue_dependencies":true},"has_wiki":true,"has_pull_requests":true,"has_projects":true,"ignore_whitespace_conflicts":false,"allow_merge_commits":true,"allow_rebase":true,"allow_rebase_explicit":true,"allow_squash_merge":true,"default_merge_style":"merge","avatar_url":"","internal":false,"mirror_interval":"","mirror_updated":"0001-01-01T00:00:00Z","repo_transfer":null}},"head":{"label":"test-user-patch-1","ref":"test-user-patch-1","sha":"4d3f126f7f6b76c01187a06ec704a8a3055591de","repo_id":1,"repo":{"id":1,"owner":{"id":1,"login":"test-user","full_name":"","email":"a@b.c","avatar_url":"http://localhost:3000/avatar/5d60d4e28066df254d5452f92c910092","language":"","is_admin":false,"last_login":"0001-01-01T00:00:00Z","created":"2022-05-15T18:42:54Z","restricted":false,"active":false,"prohibit_login":false,"location":"","website":"","description":"","visibility":"public","followers_count":0,"following_count":0,"starred_repos_count":0,"username":"test-user"},"name":"gitea-sonarqube-bot","full_name":"test-user/gitea-sonarqube-bot","description":"","empty":false,"private":false,"fork":false,"template":false,"parent":null,"mirror":false,"size":110,"html_url":"http://localhost:3000/test-user/gitea-sonarqube-bot","ssh_url":"git@localhost:test-user/gitea-sonarqube-bot.git","clone_url":"http://localhost:3000/test-user/gitea-sonarqube-bot.git","original_url":"","website":"","stars_count":0,"forks_count":0,"watchers_count":1,"open_issues_count":0,"open_pr_counter":0,"release_counter":0,"default_branch":"main","archived":false,"created_at":"2022-05-15T18:45:46Z","updated_at":"2022-05-15T18:46:09Z","permissions":{"admin":false,"push":false,"pull":true},"has_issues":true,"internal_tracker":{"enable_time_tracker":true,"allow_only_contributors_to_track_time":true,"enable_issue_dependencies":true},"has_wiki":true,"has_pull_requests":true,"has_projects":true,"ignore_whitespace_conflicts":false,"allow_merge_commits":true,"allow_rebase":true,"allow_rebase_explicit":true,"allow_squash_merge":true,"default_merge_style":"merge","avatar_url":"","internal":false,"mirror_interval":"","mirror_updated":"0001-01-01T00:00:00Z","repo_transfer":null}},"merge_base":"2e5c9f7fe85fd8fb6019b3dd299744e0afce076b","due_date":null,"created_at":"2022-05-15T18:46:19Z","updated_at":"2022-05-15T18:46:19Z","closed_at":null},"repository":{"id":1,"owner":{"id":1,"login":"test-user","full_name":"","email":"a@b.c","avatar_url":"http://localhost:3000/avatar/5d60d4e28066df254d5452f92c910092","language":"","is_admin":false,"last_login":"0001-01-01T00:00:00Z","created":"2022-05-15T18:42:54Z","restricted":false,"active":false,"prohibit_login":false,"location":"","website":"","description":"","visibility":"public","followers_count":0,"following_count":0,"starred_repos_count":0,"username":"test-user"},"name":"gitea-sonarqube-bot","full_name":"test-user/gitea-sonarqube-bot","description":"","empty":false,"private":false,"fork":false,"template":false,"parent":null,"mirror":false,"size":110,"html_url":"http://localhost:3000/test-user/gitea-sonarqube-bot","ssh_url":"git@localhost:test-user/gitea-sonarqube-bot.git","clone_url":"http://localhost:3000/test-user/gitea-sonarqube-bot.git","original_url":"","website":"","stars_count":0,"forks_count":0,"watchers_count":1,"open_issues_count":0,"open_pr_counter":0,"release_counter":0,"default_branch":"main","archived":false,"created_at":"2022-05-15T18:45:46Z","updated_at":"2022-05-15T18:46:09Z","permissions":{"admin":true,"push":true,"pull":true},"has_issues":true,"internal_tracker":{"enable_time_tracker":true,"allow_only_contributors_to_track_time":true,"enable_issue_dependencies":true},"has_wiki":true,"has_pull_requests":true,"has_projects":true,"ignore_whitespace_conflicts":false,"allow_merge_commits":true,"allow_rebase":true,"allow_rebase_explicit":true,"allow_squash_merge":true,"default_merge_style":"merge","avatar_url":"","internal":false,"mirror_interval":"","mirror_updated":"0001-01-01T00:00:00Z","repo_transfer":null},"sender":{"id":1,"login":"test-user","full_name":"","email":"a@b.c","avatar_url":"http://localhost:3000/avatar/5d60d4e28066df254d5452f92c910092","language":"","is_admin":false,"last_login":"0001-01-01T00:00:00Z","created":"2022-05-15T18:42:54Z","restricted":false,"active":false,"prohibit_login":false,"location":"","website":"","description":"","visibility":"public","followers_count":0,"following_count":0,"starred_repos_count":0,"username":"test-user"},"review":null}`))
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, `{"message": "Processing data. See bot logs for details."}`, rr.Body.String())
	})

	t.Run("With invalid JSON body", func(t *testing.T) {
		settings.Gitea = settings.GiteaConfig{
			Webhook: &settings.Webhook{
				Secret: "",
			},
		}
		settings.Projects = []settings.Project{
			{
				SonarQube: struct{ Key string }{
					Key: "gitea-sonarqube-bot",
				},
			},
		}

		req, rr, handler := withValidRequestData(t, []byte(`{ "action": ["non-string-value-for-action"] }`))
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
		assert.Equal(t, `{"message": "Error parsing POST body."}`, rr.Body.String())
	})

	t.Run("With invalid signature", func(t *testing.T) {
		settings.Gitea = settings.GiteaConfig{
			Webhook: &settings.Webhook{
				Secret: "gitea-synchronize-test-webhook",
			},
		}
		settings.Projects = []settings.Project{
			{
				SonarQube: struct{ Key string }{
					Key: "pr-bot",
				},
			},
		}
		req, rr, handler := withValidRequestData(t, []byte(`{"action":"opened","number":1,"pull_request":{"id":1,"url":"http://localhost:3000/test-user/gitea-sonarqube-bot/pulls/1","number":1,"user":{"id":1,"login":"test-user","full_name":"","email":"a@b.c","avatar_url":"http://localhost:3000/avatar/5d60d4e28066df254d5452f92c910092","language":"","is_admin":false,"last_login":"0001-01-01T00:00:00Z","created":"2022-05-15T18:42:54Z","restricted":false,"active":false,"prohibit_login":false,"location":"","website":"","description":"","visibility":"public","followers_count":0,"following_count":0,"starred_repos_count":0,"username":"test-user"},"title":"„README.md“ ändern","body":"","labels":[],"milestone":null,"assignee":null,"assignees":null,"state":"open","is_locked":false,"comments":0,"html_url":"http://localhost:3000/test-user/gitea-sonarqube-bot/pulls/1","diff_url":"http://localhost:3000/test-user/gitea-sonarqube-bot/pulls/1.diff","patch_url":"http://localhost:3000/test-user/gitea-sonarqube-bot/pulls/1.patch","mergeable":true,"merged":false,"merged_at":null,"merge_commit_sha":null,"merged_by":null,"base":{"label":"main","ref":"main","sha":"2e5c9f7fe85fd8fb6019b3dd299744e0afce076b","repo_id":1,"repo":{"id":1,"owner":{"id":1,"login":"test-user","full_name":"","email":"a@b.c","avatar_url":"http://localhost:3000/avatar/5d60d4e28066df254d5452f92c910092","language":"","is_admin":false,"last_login":"0001-01-01T00:00:00Z","created":"2022-05-15T18:42:54Z","restricted":false,"active":false,"prohibit_login":false,"location":"","website":"","description":"","visibility":"public","followers_count":0,"following_count":0,"starred_repos_count":0,"username":"test-user"},"name":"gitea-sonarqube-bot","full_name":"test-user/gitea-sonarqube-bot","description":"","empty":false,"private":false,"fork":false,"template":false,"parent":null,"mirror":false,"size":110,"html_url":"http://localhost:3000/test-user/gitea-sonarqube-bot","ssh_url":"git@localhost:test-user/gitea-sonarqube-bot.git","clone_url":"http://localhost:3000/test-user/gitea-sonarqube-bot.git","original_url":"","website":"","stars_count":0,"forks_count":0,"watchers_count":1,"open_issues_count":0,"open_pr_counter":0,"release_counter":0,"default_branch":"main","archived":false,"created_at":"2022-05-15T18:45:46Z","updated_at":"2022-05-15T18:46:09Z","permissions":{"admin":false,"push":false,"pull":true},"has_issues":true,"internal_tracker":{"enable_time_tracker":true,"allow_only_contributors_to_track_time":true,"enable_issue_dependencies":true},"has_wiki":true,"has_pull_requests":true,"has_projects":true,"ignore_whitespace_conflicts":false,"allow_merge_commits":true,"allow_rebase":true,"allow_rebase_explicit":true,"allow_squash_merge":true,"default_merge_style":"merge","avatar_url":"","internal":false,"mirror_interval":"","mirror_updated":"0001-01-01T00:00:00Z","repo_transfer":null}},"head":{"label":"test-user-patch-1","ref":"test-user-patch-1","sha":"4d3f126f7f6b76c01187a06ec704a8a3055591de","repo_id":1,"repo":{"id":1,"owner":{"id":1,"login":"test-user","full_name":"","email":"a@b.c","avatar_url":"http://localhost:3000/avatar/5d60d4e28066df254d5452f92c910092","language":"","is_admin":false,"last_login":"0001-01-01T00:00:00Z","created":"2022-05-15T18:42:54Z","restricted":false,"active":false,"prohibit_login":false,"location":"","website":"","description":"","visibility":"public","followers_count":0,"following_count":0,"starred_repos_count":0,"username":"test-user"},"name":"gitea-sonarqube-bot","full_name":"test-user/gitea-sonarqube-bot","description":"","empty":false,"private":false,"fork":false,"template":false,"parent":null,"mirror":false,"size":110,"html_url":"http://localhost:3000/test-user/gitea-sonarqube-bot","ssh_url":"git@localhost:test-user/gitea-sonarqube-bot.git","clone_url":"http://localhost:3000/test-user/gitea-sonarqube-bot.git","original_url":"","website":"","stars_count":0,"forks_count":0,"watchers_count":1,"open_issues_count":0,"open_pr_counter":0,"release_counter":0,"default_branch":"main","archived":false,"created_at":"2022-05-15T18:45:46Z","updated_at":"2022-05-15T18:46:09Z","permissions":{"admin":false,"push":false,"pull":true},"has_issues":true,"internal_tracker":{"enable_time_tracker":true,"allow_only_contributors_to_track_time":true,"enable_issue_dependencies":true},"has_wiki":true,"has_pull_requests":true,"has_projects":true,"ignore_whitespace_conflicts":false,"allow_merge_commits":true,"allow_rebase":true,"allow_rebase_explicit":true,"allow_squash_merge":true,"default_merge_style":"merge","avatar_url":"","internal":false,"mirror_interval":"","mirror_updated":"0001-01-01T00:00:00Z","repo_transfer":null}},"merge_base":"2e5c9f7fe85fd8fb6019b3dd299744e0afce076b","due_date":null,"created_at":"2022-05-15T18:46:19Z","updated_at":"2022-05-15T18:46:19Z","closed_at":null},"repository":{"id":1,"owner":{"id":1,"login":"test-user","full_name":"","email":"a@b.c","avatar_url":"http://localhost:3000/avatar/5d60d4e28066df254d5452f92c910092","language":"","is_admin":false,"last_login":"0001-01-01T00:00:00Z","created":"2022-05-15T18:42:54Z","restricted":false,"active":false,"prohibit_login":false,"location":"","website":"","description":"","visibility":"public","followers_count":0,"following_count":0,"starred_repos_count":0,"username":"test-user"},"name":"gitea-sonarqube-bot","full_name":"test-user/gitea-sonarqube-bot","description":"","empty":false,"private":false,"fork":false,"template":false,"parent":null,"mirror":false,"size":110,"html_url":"http://localhost:3000/test-user/gitea-sonarqube-bot","ssh_url":"git@localhost:test-user/gitea-sonarqube-bot.git","clone_url":"http://localhost:3000/test-user/gitea-sonarqube-bot.git","original_url":"","website":"","stars_count":0,"forks_count":0,"watchers_count":1,"open_issues_count":0,"open_pr_counter":0,"release_counter":0,"default_branch":"main","archived":false,"created_at":"2022-05-15T18:45:46Z","updated_at":"2022-05-15T18:46:09Z","permissions":{"admin":true,"push":true,"pull":true},"has_issues":true,"internal_tracker":{"enable_time_tracker":true,"allow_only_contributors_to_track_time":true,"enable_issue_dependencies":true},"has_wiki":true,"has_pull_requests":true,"has_projects":true,"ignore_whitespace_conflicts":false,"allow_merge_commits":true,"allow_rebase":true,"allow_rebase_explicit":true,"allow_squash_merge":true,"default_merge_style":"merge","avatar_url":"","internal":false,"mirror_interval":"","mirror_updated":"0001-01-01T00:00:00Z","repo_transfer":null},"sender":{"id":1,"login":"test-user","full_name":"","email":"a@b.c","avatar_url":"http://localhost:3000/avatar/5d60d4e28066df254d5452f92c910092","language":"","is_admin":false,"last_login":"0001-01-01T00:00:00Z","created":"2022-05-15T18:42:54Z","restricted":false,"active":false,"prohibit_login":false,"location":"","website":"","description":"","visibility":"public","followers_count":0,"following_count":0,"starred_repos_count":0,"username":"test-user"},"review":null}`))
		req.Header.Set("X-Gitea-Signature", "647f2395d30b1b7efcb58d9338be5b69c2addb54faf6bde6314a57ea28f45467")
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusPreconditionFailed, rr.Code)
		assert.Equal(t, `{"message": "Webhook validation failed. Request rejected."}`, rr.Body.String())
	})

	t.Run("With ignored project", func(t *testing.T) {
		settings.Gitea = settings.GiteaConfig{
			Webhook: &settings.Webhook{
				Secret: "",
			},
		}
		settings.Projects = []settings.Project{
			{
				SonarQube: struct{ Key string }{
					Key: "gitea-sonarqube-bot",
				},
			},
		}
		req, rr, handler := withValidRequestData(t, []byte(`{"action":"opened","number":1,"pull_request":{"id":1,"url":"http://localhost:3000/test-user/gitea-sonarqube-bot/pulls/1","number":1,"user":{"id":1,"login":"test-user","full_name":"","email":"a@b.c","avatar_url":"http://localhost:3000/avatar/5d60d4e28066df254d5452f92c910092","language":"","is_admin":false,"last_login":"0001-01-01T00:00:00Z","created":"2022-05-15T18:42:54Z","restricted":false,"active":false,"prohibit_login":false,"location":"","website":"","description":"","visibility":"public","followers_count":0,"following_count":0,"starred_repos_count":0,"username":"test-user"},"title":"„README.md“ ändern","body":"","labels":[],"milestone":null,"assignee":null,"assignees":null,"state":"open","is_locked":false,"comments":0,"html_url":"http://localhost:3000/test-user/gitea-sonarqube-bot/pulls/1","diff_url":"http://localhost:3000/test-user/gitea-sonarqube-bot/pulls/1.diff","patch_url":"http://localhost:3000/test-user/gitea-sonarqube-bot/pulls/1.patch","mergeable":true,"merged":false,"merged_at":null,"merge_commit_sha":null,"merged_by":null,"base":{"label":"main","ref":"main","sha":"2e5c9f7fe85fd8fb6019b3dd299744e0afce076b","repo_id":1,"repo":{"id":1,"owner":{"id":1,"login":"test-user","full_name":"","email":"a@b.c","avatar_url":"http://localhost:3000/avatar/5d60d4e28066df254d5452f92c910092","language":"","is_admin":false,"last_login":"0001-01-01T00:00:00Z","created":"2022-05-15T18:42:54Z","restricted":false,"active":false,"prohibit_login":false,"location":"","website":"","description":"","visibility":"public","followers_count":0,"following_count":0,"starred_repos_count":0,"username":"test-user"},"name":"gitea-sonarqube-bot","full_name":"test-user/gitea-sonarqube-bot","description":"","empty":false,"private":false,"fork":false,"template":false,"parent":null,"mirror":false,"size":110,"html_url":"http://localhost:3000/test-user/gitea-sonarqube-bot","ssh_url":"git@localhost:test-user/gitea-sonarqube-bot.git","clone_url":"http://localhost:3000/test-user/gitea-sonarqube-bot.git","original_url":"","website":"","stars_count":0,"forks_count":0,"watchers_count":1,"open_issues_count":0,"open_pr_counter":0,"release_counter":0,"default_branch":"main","archived":false,"created_at":"2022-05-15T18:45:46Z","updated_at":"2022-05-15T18:46:09Z","permissions":{"admin":false,"push":false,"pull":true},"has_issues":true,"internal_tracker":{"enable_time_tracker":true,"allow_only_contributors_to_track_time":true,"enable_issue_dependencies":true},"has_wiki":true,"has_pull_requests":true,"has_projects":true,"ignore_whitespace_conflicts":false,"allow_merge_commits":true,"allow_rebase":true,"allow_rebase_explicit":true,"allow_squash_merge":true,"default_merge_style":"merge","avatar_url":"","internal":false,"mirror_interval":"","mirror_updated":"0001-01-01T00:00:00Z","repo_transfer":null}},"head":{"label":"test-user-patch-1","ref":"test-user-patch-1","sha":"4d3f126f7f6b76c01187a06ec704a8a3055591de","repo_id":1,"repo":{"id":1,"owner":{"id":1,"login":"test-user","full_name":"","email":"a@b.c","avatar_url":"http://localhost:3000/avatar/5d60d4e28066df254d5452f92c910092","language":"","is_admin":false,"last_login":"0001-01-01T00:00:00Z","created":"2022-05-15T18:42:54Z","restricted":false,"active":false,"prohibit_login":false,"location":"","website":"","description":"","visibility":"public","followers_count":0,"following_count":0,"starred_repos_count":0,"username":"test-user"},"name":"gitea-sonarqube-bot","full_name":"test-user/gitea-sonarqube-bot","description":"","empty":false,"private":false,"fork":false,"template":false,"parent":null,"mirror":false,"size":110,"html_url":"http://localhost:3000/test-user/gitea-sonarqube-bot","ssh_url":"git@localhost:test-user/gitea-sonarqube-bot.git","clone_url":"http://localhost:3000/test-user/gitea-sonarqube-bot.git","original_url":"","website":"","stars_count":0,"forks_count":0,"watchers_count":1,"open_issues_count":0,"open_pr_counter":0,"release_counter":0,"default_branch":"main","archived":false,"created_at":"2022-05-15T18:45:46Z","updated_at":"2022-05-15T18:46:09Z","permissions":{"admin":false,"push":false,"pull":true},"has_issues":true,"internal_tracker":{"enable_time_tracker":true,"allow_only_contributors_to_track_time":true,"enable_issue_dependencies":true},"has_wiki":true,"has_pull_requests":true,"has_projects":true,"ignore_whitespace_conflicts":false,"allow_merge_commits":true,"allow_rebase":true,"allow_rebase_explicit":true,"allow_squash_merge":true,"default_merge_style":"merge","avatar_url":"","internal":false,"mirror_interval":"","mirror_updated":"0001-01-01T00:00:00Z","repo_transfer":null}},"merge_base":"2e5c9f7fe85fd8fb6019b3dd299744e0afce076b","due_date":null,"created_at":"2022-05-15T18:46:19Z","updated_at":"2022-05-15T18:46:19Z","closed_at":null},"repository":{"id":1,"owner":{"id":1,"login":"test-user","full_name":"","email":"a@b.c","avatar_url":"http://localhost:3000/avatar/5d60d4e28066df254d5452f92c910092","language":"","is_admin":false,"last_login":"0001-01-01T00:00:00Z","created":"2022-05-15T18:42:54Z","restricted":false,"active":false,"prohibit_login":false,"location":"","website":"","description":"","visibility":"public","followers_count":0,"following_count":0,"starred_repos_count":0,"username":"test-user"},"name":"gitea-sonarqube-bot","full_name":"test-user/gitea-sonarqube-bot","description":"","empty":false,"private":false,"fork":false,"template":false,"parent":null,"mirror":false,"size":110,"html_url":"http://localhost:3000/test-user/gitea-sonarqube-bot","ssh_url":"git@localhost:test-user/gitea-sonarqube-bot.git","clone_url":"http://localhost:3000/test-user/gitea-sonarqube-bot.git","original_url":"","website":"","stars_count":0,"forks_count":0,"watchers_count":1,"open_issues_count":0,"open_pr_counter":0,"release_counter":0,"default_branch":"main","archived":false,"created_at":"2022-05-15T18:45:46Z","updated_at":"2022-05-15T18:46:09Z","permissions":{"admin":true,"push":true,"pull":true},"has_issues":true,"internal_tracker":{"enable_time_tracker":true,"allow_only_contributors_to_track_time":true,"enable_issue_dependencies":true},"has_wiki":true,"has_pull_requests":true,"has_projects":true,"ignore_whitespace_conflicts":false,"allow_merge_commits":true,"allow_rebase":true,"allow_rebase_explicit":true,"allow_squash_merge":true,"default_merge_style":"merge","avatar_url":"","internal":false,"mirror_interval":"","mirror_updated":"0001-01-01T00:00:00Z","repo_transfer":null},"sender":{"id":1,"login":"test-user","full_name":"","email":"a@b.c","avatar_url":"http://localhost:3000/avatar/5d60d4e28066df254d5452f92c910092","language":"","is_admin":false,"last_login":"0001-01-01T00:00:00Z","created":"2022-05-15T18:42:54Z","restricted":false,"active":false,"prohibit_login":false,"location":"","website":"","description":"","visibility":"public","followers_count":0,"following_count":0,"starred_repos_count":0,"username":"test-user"},"review":null}`))
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, `{"message": "ignore hook for non-configured project 'test-user/gitea-sonarqube-bot'"}`, rr.Body.String())
	})
}
