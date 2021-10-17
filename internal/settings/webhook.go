package settings

import (
	"fmt"
	"io/ioutil"
)

type webhook struct {
	Secret     string
	secretFile string
}

func (w *webhook) lookupSecret(errCallback func(string)) {
	if w.secretFile == "" {
		return
	}

	content, err := ioutil.ReadFile(w.secretFile)
	if err != nil {
		errCallback(fmt.Sprintf("Cannot read '%s' or it is no regular file: %s", w.secretFile, err.Error()))
		return
	}

	w.Secret = string(content)
}

func NewWebhook(extractor func(string) string, confContainer string, errCallback func(string)) *webhook {
	w := &webhook{
		Secret:     extractor(fmt.Sprintf("%s.webhook.secret", confContainer)),
		secretFile: extractor(fmt.Sprintf("%s.webhook.secretFile", confContainer)),
	}

	w.lookupSecret(errCallback)

	return w
}
