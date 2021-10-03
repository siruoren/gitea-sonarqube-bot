package webhook_handler

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
)

// SETUP: mute logs
func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}
