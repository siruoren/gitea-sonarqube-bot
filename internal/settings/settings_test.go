package settings

import (
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

var defaultConfig []byte = []byte(
	`gitea:
  url: https://example.com/gitea
  token:
    value: d0fcdeb5eaa99c506831f9eb4e63fc7cc484a565
  webhook:
    secret: haxxor-gitea-secret
sonarqube:
  url: https://example.com/sonarqube
  token:
    value: a09eb5785b25bb2cbacf48808a677a0709f02d8e
  webhook:
    secret: haxxor-sonarqube-secret
  additionalMetrics: []
projects:
  - sonarqube:
      key: gitea-sonarqube-pr-bot
    gitea:
      owner: example-organization
      name: pr-bot
namingPattern:
  regex: "^PR-(\\d+)$"
  template: "PR-%d"
`)

func WriteConfigFile(t *testing.T, content []byte) string {
	dir := os.TempDir()
	config := path.Join(dir, "config.yaml")

	t.Cleanup(func() {
		os.Remove(config)
	})

	_ = ioutil.WriteFile(config, content, 0444)

	return config
}

func TestLoadWithMissingFile(t *testing.T) {
	assert.Panics(t, func() { Load(path.Join(os.TempDir(), "config.yaml")) }, "No panic while reading missing file")
}

func TestLoadWithExistingFile(t *testing.T) {
	c := WriteConfigFile(t, defaultConfig)

	assert.NotPanics(t, func() { Load(c) }, "Unexpected panic while reading existing file")
}

func TestLoadGiteaStructure(t *testing.T) {
	c := WriteConfigFile(t, defaultConfig)
	Load(c)

	expected := GiteaConfig{
		Url: "https://example.com/gitea",
		Token: &Token{
			Value: "d0fcdeb5eaa99c506831f9eb4e63fc7cc484a565",
		},
		Webhook: &Webhook{
			Secret: "haxxor-gitea-secret",
		},
	}

	assert.EqualValues(t, expected, Gitea)
}

func TestLoadGiteaStructureInjectedEnvs(t *testing.T) {
	os.Setenv("PRBOT_GITEA_WEBHOOK_SECRET", "injected-webhook-secret")
	os.Setenv("PRBOT_GITEA_TOKEN_VALUE", "injected-token")
	c := WriteConfigFile(t, defaultConfig)
	Load(c)

	expected := GiteaConfig{
		Url: "https://example.com/gitea",
		Token: &Token{
			Value: "injected-token",
		},
		Webhook: &Webhook{
			Secret: "injected-webhook-secret",
		},
	}

	assert.EqualValues(t, expected, Gitea)

	t.Cleanup(func() {
		os.Unsetenv("PRBOT_GITEA_WEBHOOK_SECRET")
		os.Unsetenv("PRBOT_GITEA_TOKEN_VALUE")
	})
}

func TestLoadSonarQubeStructure(t *testing.T) {
	c := WriteConfigFile(t, defaultConfig)
	Load(c)

	expected := SonarQubeConfig{
		Url: "https://example.com/sonarqube",
		Token: &Token{
			Value: "a09eb5785b25bb2cbacf48808a677a0709f02d8e",
		},
		Webhook: &Webhook{
			Secret: "haxxor-sonarqube-secret",
		},
	}

	assert.EqualValues(t, expected, SonarQube)
	assert.EqualValues(t, expected.GetMetricsList(), "bugs,vulnerabilities,code_smells")
}

func TestLoadSonarQubeStructureWithAdditionalMetrics(t *testing.T) {
	c := WriteConfigFile(t, []byte(
		`gitea:
  url: https://example.com/gitea
  token:
    value: fake-gitea-token
sonarqube:
  url: https://example.com/sonarqube
  token:
    value: fake-sonarqube-token
  additionalMetrics: "new_security_hotspots"
projects:
  - sonarqube:
      key: gitea-sonarqube-pr-bot
    gitea:
      owner: example-organization
      name: pr-bot
`))
	Load(c)

	expected := SonarQubeConfig{
		Url: "https://example.com/sonarqube",
		Token: &Token{
			Value: "fake-sonarqube-token",
		},
		Webhook: &Webhook{
			Secret: "",
		},
		AdditionalMetrics: []string{
			"new_security_hotspots",
		},
	}

	assert.EqualValues(t, expected, SonarQube)
	assert.EqualValues(t, expected.AdditionalMetrics, []string{"new_security_hotspots"})
	assert.EqualValues(t, "bugs,vulnerabilities,code_smells,new_security_hotspots", SonarQube.GetMetricsList())
}

func TestLoadSonarQubeStructureInjectedEnvs(t *testing.T) {
	os.Setenv("PRBOT_SONARQUBE_WEBHOOK_SECRET", "injected-webhook-secret")
	os.Setenv("PRBOT_SONARQUBE_TOKEN_VALUE", "injected-token")
	c := WriteConfigFile(t, defaultConfig)
	Load(c)

	expected := SonarQubeConfig{
		Url: "https://example.com/sonarqube",
		Token: &Token{
			Value: "injected-token",
		},
		Webhook: &Webhook{
			Secret: "injected-webhook-secret",
		},
	}

	assert.EqualValues(t, expected, SonarQube)

	t.Cleanup(func() {
		os.Unsetenv("PRBOT_SONARQUBE_WEBHOOK_SECRET")
		os.Unsetenv("PRBOT_SONARQUBE_TOKEN_VALUE")
	})
}

func TestLoadStructureWithFileReferenceResolving(t *testing.T) {
	giteaWebhookSecretFile := path.Join(os.TempDir(), "webhook-secret-gitea")
	_ = ioutil.WriteFile(giteaWebhookSecretFile, []byte(`gitea-totally-secret`), 0444)

	giteaTokenFile := path.Join(os.TempDir(), "token-secret-gitea")
	_ = ioutil.WriteFile(giteaTokenFile, []byte(`d0fcdeb5eaa99c506831f9eb4e63fc7cc484a565`), 0444)

	sonarqubeWebhookSecretFile := path.Join(os.TempDir(), "webhook-secret-sonarqube")
	_ = ioutil.WriteFile(sonarqubeWebhookSecretFile, []byte(`sonarqube-totally-secret`), 0444)

	sonarqubeTokenFile := path.Join(os.TempDir(), "token-secret-sonarqube")
	_ = ioutil.WriteFile(sonarqubeTokenFile, []byte(`a09eb5785b25bb2cbacf48808a677a0709f02d8e`), 0444)

	c := WriteConfigFile(t, []byte(
		`gitea:
  url: https://example.com/gitea
  token:
    value: fake-gitea-token
sonarqube:
  url: https://example.com/sonarqube
  token:
    value: fake-sonarqube-token
projects:
  - sonarqube:
      key: gitea-sonarqube-pr-bot
    gitea:
      owner: example-organization
      name: pr-bot
`))
	os.Setenv("PRBOT_GITEA_WEBHOOK_SECRETFILE", giteaWebhookSecretFile)
	os.Setenv("PRBOT_GITEA_TOKEN_FILE", giteaTokenFile)
	os.Setenv("PRBOT_SONARQUBE_WEBHOOK_SECRETFILE", sonarqubeWebhookSecretFile)
	os.Setenv("PRBOT_SONARQUBE_TOKEN_FILE", sonarqubeTokenFile)

	expectedGitea := GiteaConfig{
		Url: "https://example.com/gitea",
		Token: &Token{
			Value: "d0fcdeb5eaa99c506831f9eb4e63fc7cc484a565",
			file:  giteaTokenFile,
		},
		Webhook: &Webhook{
			Secret:     "gitea-totally-secret",
			secretFile: giteaWebhookSecretFile,
		},
	}

	expectedSonarQube := SonarQubeConfig{
		Url: "https://example.com/sonarqube",
		Token: &Token{
			Value: "a09eb5785b25bb2cbacf48808a677a0709f02d8e",
			file:  sonarqubeTokenFile,
		},
		Webhook: &Webhook{
			Secret:     "sonarqube-totally-secret",
			secretFile: sonarqubeWebhookSecretFile,
		},
		AdditionalMetrics: []string{},
	}

	Load(c)
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

func TestLoadProjectsStructure(t *testing.T) {
	c := WriteConfigFile(t, defaultConfig)
	Load(c)

	expectedProjects := []Project{
		{
			SonarQube: struct{ Key string }{
				Key: "gitea-sonarqube-pr-bot",
			},
			Gitea: GiteaRepository{
				Owner: "example-organization",
				Name:  "pr-bot",
			},
		},
	}

	assert.EqualValues(t, expectedProjects, Projects)
}

func TestLoadProjectsStructureWithNoMapping(t *testing.T) {
	invalidConfig := []byte(
		`gitea:
  url: https://example.com/gitea
  token:
    value: d0fcdeb5eaa99c506831f9eb4e63fc7cc484a565
  webhook:
    secret: haxxor-gitea-secret
sonarqube:
  url: https://example.com/sonarqube
  token:
    value: a09eb5785b25bb2cbacf48808a677a0709f02d8e
  webhook:
    secret: haxxor-sonarqube-secret
projects: []
`)
	c := WriteConfigFile(t, invalidConfig)

	assert.Panics(t, func() { Load(c) }, "No panic for empty project mapping that is required")
}

func TestLoadNamingPatternStructure(t *testing.T) {
	c := WriteConfigFile(t, defaultConfig)
	Load(c)

	expected := &PatternConfig{
		RegExp:   regexp.MustCompile(`^PR-(\d+)$`),
		Template: "PR-%d",
	}

	assert.EqualValues(t, expected, Pattern)
}

func TestLoadNamingPatternStructureWithInternalDefaults(t *testing.T) {
	c := WriteConfigFile(t, []byte(
		`gitea:
  url: https://example.com/gitea
  token:
    value: fake-gitea-token
sonarqube:
  url: https://example.com/sonarqube
  token:
    value: fake-sonarqube-token
  additionalMetrics: "new_security_hotspots"
projects:
  - sonarqube:
      key: gitea-sonarqube-pr-bot
    gitea:
      owner: example-organization
      name: pr-bot
`))
	Load(c)

	expected := &PatternConfig{
		RegExp:   regexp.MustCompile(`^PR-(\d+)$`),
		Template: "PR-%d",
	}

	assert.EqualValues(t, expected, Pattern)
}

func TestLoadNamingPatternStructureInjectedEnvs(t *testing.T) {
	os.Setenv("PRBOT_NAMINGPATTERN_REGEX", "test-(\\d+)-pullrequest")
	os.Setenv("PRBOT_NAMINGPATTERN_TEMPLATE", "test-%d-pullrequest")
	c := WriteConfigFile(t, defaultConfig)
	Load(c)

	expected := &PatternConfig{
		RegExp:   regexp.MustCompile(`test-(\d+)-pullrequest`),
		Template: "test-%d-pullrequest",
	}

	assert.EqualValues(t, expected, Pattern)

	t.Cleanup(func() {
		os.Unsetenv("PRBOT_NAMINGPATTERN_REGEX")
		os.Unsetenv("PRBOT_NAMINGPATTERN_TEMPLATE")
	})
}

func TestLoadNamingPatternStructureMixedInput(t *testing.T) {
	os.Setenv("PRBOT_NAMINGPATTERN_REGEX", "test-(\\d+)-pullrequest")
	c := WriteConfigFile(t, defaultConfig)
	Load(c)

	expected := &PatternConfig{
		RegExp:   regexp.MustCompile(`test-(\d+)-pullrequest`),
		Template: "PR-%d",
	}

	assert.EqualValues(t, expected, Pattern)

	t.Cleanup(func() {
		os.Unsetenv("PRBOT_NAMINGPATTERN_REGEX")
	})
}
