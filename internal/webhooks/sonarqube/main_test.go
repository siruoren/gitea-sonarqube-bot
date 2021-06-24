package sonarqube

import (
	"log"
	"io/ioutil"
	"os"
	"testing"
)

// SETUP: mute logs
func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}
