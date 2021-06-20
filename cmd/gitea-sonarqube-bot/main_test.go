package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfigLocationWithDefault(t *testing.T) {
	assert.Equal(t, "config", GetConfigLocation())
}

func TestGetConfigLocationWithEnvironmentOverride(t *testing.T) {
	os.Setenv("PRBOT_CONFIG_PATH", "/tmp/")

	assert.Equal(t, "/tmp/", GetConfigLocation())

	t.Cleanup(func() {
		os.Unsetenv("PRBOT_CONFIG_PATH")
	})
}
