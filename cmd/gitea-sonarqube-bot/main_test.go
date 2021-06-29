package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfigLocationWithDefault(t *testing.T) {
	assert.Equal(t, "config", getConfigLocation())
}

func TestGetConfigLocationWithEnvironmentOverride(t *testing.T) {
	os.Setenv("PRBOT_CONFIG_PATH", "/tmp/")

	assert.Equal(t, "/tmp/", getConfigLocation())

	t.Cleanup(func() {
		os.Unsetenv("PRBOT_CONFIG_PATH")
	})
}
