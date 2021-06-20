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
  token: d0fcdeb5eaa99c506831f9eb4e63fc7cc484a565
  webhookSecret:
    value: haxxor-gitea-secret
  repositories: []
sonarqube:
  url: https://example.com/sonarqube
  token: a09eb5785b25bb2cbacf48808a677a0709f02d8e
  webhookSecret:
    value: haxxor-sonarqube-secret
  projects: []
`)

var incompleteConfig []byte = []byte(
`gitea:
  url: https://example.com/gitea
  webhookSecret:
    value: haxxor-gitea-secret
sonarqube:
  url: https://example.com/sonarqube
  token: a09eb5785b25bb2cbacf48808a677a0709f02d8e
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
		Token: "d0fcdeb5eaa99c506831f9eb4e63fc7cc484a565",
		WebhookSecret: WebhookSecret{
			Value: "haxxor-gitea-secret",
		},
		Repositories: []GiteaRepository{},
	}

	assert.EqualValues(t, expected, Gitea)
}

func TestLoadGiteaStructureWithEnvInjectedWebhookSecret(t *testing.T) {
	os.Setenv("PRBOT_GITEA_WEBHOOKSECRET_VALUE", "injected-secret")
	os.Setenv("PRBOT_GITEA_TOKEN", "injected-token")
	WriteConfigFile(t, defaultConfigInlineSecrets)
	Load(os.TempDir())

	expected := GiteaConfig{
		Url: "https://example.com/gitea",
		Token: "injected-token",
		WebhookSecret: WebhookSecret{
			Value: "injected-secret",
		},
		Repositories: []GiteaRepository{},
	}

	assert.EqualValues(t, expected, Gitea)

	t.Cleanup(func() {
		os.Unsetenv("PRBOT_GITEA_WEBHOOKSECRET_VALUE")
		os.Unsetenv("PRBOT_GITEA_TOKEN")
	})
}

func TestLoadStructureWithResolvedWebhookFileFromEnvInjected(t *testing.T) {
	secretFile := path.Join(os.TempDir(), "webhook-secret-sonarqube")
	_ = ioutil.WriteFile(secretFile, []byte(`totally-secret`),0444)

	os.Setenv("PRBOT_GITEA_WEBHOOKSECRET_FILE", secretFile)
	os.Setenv("PRBOT_SONARQUBE_WEBHOOKSECRET_FILE", secretFile)
	os.Setenv("PRBOT_GITEA_TOKEN", "injected-token")

	WriteConfigFile(t, incompleteConfig)
	Load(os.TempDir())

	expectedGitea := GiteaConfig{
		Url: "https://example.com/gitea",
		Token: "injected-token",
		WebhookSecret: WebhookSecret{
			Value: "totally-secret",
			File: secretFile,
		},
		Repositories: []GiteaRepository{},
	}

	expectedSonarQube := SonarQubeConfig{
		Url: "https://example.com/sonarqube",
		Token: "a09eb5785b25bb2cbacf48808a677a0709f02d8e",
		WebhookSecret: WebhookSecret{
			Value: "totally-secret",
			File: secretFile,
		},
		Projects: []string{},
	}

	assert.EqualValues(t, expectedGitea, Gitea)
	assert.EqualValues(t, expectedSonarQube, SonarQube)

	t.Cleanup(func() {
		os.Remove(secretFile)
		os.Unsetenv("PRBOT_SONARQUBE_WEBHOOKSECRET_FILE")
		os.Unsetenv("PRBOT_GITEA_TOKEN")
	})
}

func TestReadSecretFileWhenDirectoryProvided(t *testing.T) {
	assert.Panics(t, func() { ReadSecretFile(os.TempDir()) }, "No panic while trying to read content from directory")
}

func TestReadSecretFileWhenMissingFileProvided(t *testing.T) {
	assert.Panics(t, func() { ReadSecretFile(path.Join(os.TempDir(), "secret-file")) }, "No panic while trying to read missing file")
}

func TestReadSecretFile(t *testing.T) {
	secretFile := path.Join(os.TempDir(), "secret-file")
	_ = ioutil.WriteFile(secretFile, []byte(`awesome-secret-content`),0444)

	assert.Equal(t, "awesome-secret-content", ReadSecretFile(secretFile))

	t.Cleanup(func() {
		os.Remove(secretFile)
	})
}
