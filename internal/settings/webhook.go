package settings

import (
	"fmt"
	"io/ioutil"

	"github.com/spf13/viper"
)

type webhook struct {
	Secret string
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

func NewWebhook(v *viper.Viper, confContainer string, errCallback func(string)) *webhook {
	w := &webhook{
		Secret: v.GetString(fmt.Sprintf("%s.webhook.secret", confContainer)),
		secretFile: v.GetString(fmt.Sprintf("%s.webhook.secretFile", confContainer)),
	}

	w.lookupSecret(errCallback)

	return w
}
