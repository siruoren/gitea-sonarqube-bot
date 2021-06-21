package settings

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

var defaultConfigInlineSecrets []byte = []byte(
`gitea:
  url: https://example.com/gitea
  token:
    value: d0fcdeb5eaa99c506831f9eb4e63fc7cc484a565
  webhook:
    secret: haxxor-gitea-secret
  repositories:
    - owner: some-owner
      name: a-repository-name
sonarqube:
  url: https://example.com/sonarqube
  token:
    value: a09eb5785b25bb2cbacf48808a677a0709f02d8e
  webhook:
    secret: haxxor-sonarqube-secret
  projects: []
`)

func WriteConfigFile(t *testing.T, content []byte) {
	dir := os.TempDir()
	config := path.Join(dir, "config.yaml")

	t.Cleanup(func() {
		os.Remove(config)
	})

	_ = ioutil.WriteFile(config, content,0444)
}

func TestLoadWithMissingFile(t *testing.T) {
	assert.Panics(t, func() { Load(os.TempDir()) }, "No panic while reading missing file")
}

func TestLoadWithExistingFile(t *testing.T) {
	WriteConfigFile(t, defaultConfigInlineSecrets)

	assert.NotPanics(t, func() { Load(os.TempDir()) }, "Unexpected panic while reading existing file")
}

func TestLoadGiteaStructure(t *testing.T) {
	WriteConfigFile(t, defaultConfigInlineSecrets)
	Load(os.TempDir())

	expected := GiteaConfig{
		Url: "https://example.com/gitea",
		Token: Token{
			Value: "d0fcdeb5eaa99c506831f9eb4e63fc7cc484a565",
		},
		Webhook: Webhook{
			Secret: "haxxor-gitea-secret",
		},
		Repositories: []GiteaRepository{
			GiteaRepository{
				Owner: "some-owner",
				Name: "a-repository-name",
			},
		},
	}

	assert.EqualValues(t, expected, Gitea)
}

func TestLoadGiteaStructureInjectedEnvs(t *testing.T) {
	os.Setenv("PRBOT_GITEA_WEBHOOK_SECRET", "injected-webhook-secret")
	os.Setenv("PRBOT_GITEA_TOKEN_VALUE", "injected-token")
	WriteConfigFile(t, defaultConfigInlineSecrets)
	Load(os.TempDir())

	expected := GiteaConfig{
		Url: "https://example.com/gitea",
		Token: Token{
			Value: "injected-token",
		},
		Webhook: Webhook{
			Secret: "injected-webhook-secret",
		},
		Repositories: []GiteaRepository{
			GiteaRepository{
				Owner: "some-owner",
				Name: "a-repository-name",
			},
		},
	}

	assert.EqualValues(t, expected, Gitea)

	t.Cleanup(func() {
		os.Unsetenv("PRBOT_GITEA_WEBHOOK_SECRET")
		os.Unsetenv("PRBOT_GITEA_TOKEN_VALUE")
	})
}

func TestLoadSonarQubeStructure(t *testing.T) {
	WriteConfigFile(t, defaultConfigInlineSecrets)
	Load(os.TempDir())

	expected := SonarQubeConfig{
		Url: "https://example.com/sonarqube",
		Token: Token{
			Value: "a09eb5785b25bb2cbacf48808a677a0709f02d8e",
		},
		Webhook: Webhook{
			Secret: "haxxor-sonarqube-secret",
		},
		Projects: []string{},
	}

	assert.EqualValues(t, expected, SonarQube)
}

func TestLoadSonarQubeStructureInjectedEnvs(t *testing.T) {
	os.Setenv("PRBOT_SONARQUBE_WEBHOOK_SECRET", "injected-webhook-secret")
	os.Setenv("PRBOT_SONARQUBE_TOKEN_VALUE", "injected-token")
	WriteConfigFile(t, defaultConfigInlineSecrets)
	Load(os.TempDir())

	expected := SonarQubeConfig{
		Url: "https://example.com/sonarqube",
		Token: Token{
			Value: "injected-token",
		},
		Webhook: Webhook{
			Secret: "injected-webhook-secret",
		},
		Projects: []string{},
	}

	assert.EqualValues(t, expected, SonarQube)

	t.Cleanup(func() {
		os.Unsetenv("PRBOT_SONARQUBE_WEBHOOK_SECRET")
		os.Unsetenv("PRBOT_SONARQUBE_TOKEN_VALUE")
	})
}

func TestLoadStructureWithFileReferenceResolving(t *testing.T) {
	giteaWebhookSecretFile := path.Join(os.TempDir(), "webhook-secret-gitea")
	_ = ioutil.WriteFile(giteaWebhookSecretFile, []byte(`gitea-totally-secret`),0444)

	giteaTokenFile := path.Join(os.TempDir(), "token-secret-gitea")
	_ = ioutil.WriteFile(giteaTokenFile, []byte(`d0fcdeb5eaa99c506831f9eb4e63fc7cc484a565`),0444)

	sonarqubeWebhookSecretFile := path.Join(os.TempDir(), "webhook-secret-sonarqube")
	_ = ioutil.WriteFile(sonarqubeWebhookSecretFile, []byte(`sonarqube-totally-secret`),0444)

	sonarqubeTokenFile := path.Join(os.TempDir(), "token-secret-sonarqube")
	_ = ioutil.WriteFile(sonarqubeTokenFile, []byte(`a09eb5785b25bb2cbacf48808a677a0709f02d8e`),0444)

	WriteConfigFile(t, []byte(
`gitea:
  url: https://example.com/gitea
  token:
    value: fake-gitea-token
  repositories: []
sonarqube:
  url: https://example.com/sonarqube
  token:
    value: fake-sonarqube-token
  projects: []
`))
	os.Setenv("PRBOT_GITEA_WEBHOOK_SECRETFILE", giteaWebhookSecretFile)
	os.Setenv("PRBOT_GITEA_TOKEN_FILE", giteaTokenFile)
	os.Setenv("PRBOT_SONARQUBE_WEBHOOK_SECRETFILE", sonarqubeWebhookSecretFile)
	os.Setenv("PRBOT_SONARQUBE_TOKEN_FILE", sonarqubeTokenFile)

	expectedGitea := GiteaConfig{
		Url: "https://example.com/gitea",
		Token: Token{
			Value: "d0fcdeb5eaa99c506831f9eb4e63fc7cc484a565",
			File: giteaTokenFile,
		},
		Webhook: Webhook{
			Secret: "gitea-totally-secret",
			SecretFile: giteaWebhookSecretFile,
		},
		Repositories: []GiteaRepository{},
	}

	expectedSonarQube := SonarQubeConfig{
		Url: "https://example.com/sonarqube",
		Token: Token{
			Value: "a09eb5785b25bb2cbacf48808a677a0709f02d8e",
			File: sonarqubeTokenFile,
		},
		Webhook: Webhook{
			Secret: "sonarqube-totally-secret",
			SecretFile: sonarqubeWebhookSecretFile,
		},
		Projects: []string{},
	}

	Load(os.TempDir())
	assert.EqualValues(t, expectedGitea, Gitea)
	assert.EqualValues(t, expectedSonarQube, SonarQube)

	t.Cleanup(func() {
		os.Remove(giteaWebhookSecretFile)
		os.Remove(giteaTokenFile)
		os.Remove(sonarqubeWebhookSecretFile)
		os.Remove(sonarqubeTokenFile)
		os.Unsetenv("PRBOT_GITEA_WEBHOOK_SECRETFILE")
		os.Unsetenv("PRBOT_GITEA_TOKEN_FILE")
		os.Unsetenv("PRBOT_SONARQUBE_WEBHOOK_SECRETFILE")
		os.Unsetenv("PRBOT_SONARQUBE_TOKEN_FILE")
	})
}
