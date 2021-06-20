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
  webhookSecret: "haxxor"
  repositories: []
sonarqube:
  url: https://example.com/sonarqube
  token: a09eb5785b25bb2cbacf48808a677a0709f02d8e
  webhookSecret: "haxxor"
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

func TestLoadWithMissingGiteaStructure(t *testing.T) {
	WriteConfigFile(t, []byte(``))

	assert.Panics(t, func() { Load(os.TempDir()) }, "No panic when Gitea is not configured")
}

func TestLoadGiteaStructure(t *testing.T) {
	WriteConfigFile(t, defaultConfigInlineSecrets)
	Load(os.TempDir())

	expected := GiteaConfig{
		Url: "https://example.com/gitea",
		Token: "d0fcdeb5eaa99c506831f9eb4e63fc7cc484a565",
		WebhookSecret: "haxxor",
		Repositories: []GiteaRepository{},
	}

	assert.EqualValues(t, expected, Gitea)
}

func TestLoadGiteaStructureWithEnvInjectedWebhookSecret(t *testing.T) {
	os.Setenv("PRBOT_GITEA_WEBHOOKSECRET", "injected-secret")
	os.Setenv("PRBOT_GITEA_TOKEN", "injected-token")
	WriteConfigFile(t, defaultConfigInlineSecrets)
	Load(os.TempDir())

	expected := GiteaConfig{
		Url: "https://example.com/gitea",
		Token: "injected-token",
		WebhookSecret: "injected-secret",
		Repositories: []GiteaRepository{},
	}

	assert.EqualValues(t, expected, Gitea)

	t.Cleanup(func() {
		os.Unsetenv("PRBOT_GITEA_WEBHOOKSECRET")
		os.Unsetenv("PRBOT_GITEA_TOKEN")
	})
}
