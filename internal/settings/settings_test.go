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
  token: 1337
  webhookSecret: ""
  repositories: []
sonarqube:
  url: https://example.com/sonarqube
  token: 42
  webhookSecret: ""
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

	expected := &GiteaConfig{
		Url: "https://example.com/gitea",
		Token: "1337",
		WebhookSecret: "",
		Repositories: []GiteaRepository{},
	}

	assert.True(t, assert.ObjectsAreEqualValues(&Gitea, expected))
}
