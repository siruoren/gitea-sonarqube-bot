package settings

import (
	"fmt"
	"io/ioutil"
)

type Token struct {
	Value string
	file  string
}

func (t *Token) lookupSecret(errCallback func(string)) {
	if t.file == "" {
		return
	}

	content, err := ioutil.ReadFile(t.file)
	if err != nil {
		errCallback(fmt.Sprintf("Cannot read '%s' or it is no regular file: %s", t.file, err.Error()))
		return
	}

	t.Value = string(content)
}

func NewToken(extractor func(string) string, confContainer string, errCallback func(string)) *Token {
	t := &Token{
		Value: extractor(fmt.Sprintf("%s.token.value", confContainer)),
		file:  extractor(fmt.Sprintf("%s.token.file", confContainer)),
	}

	t.lookupSecret(errCallback)

	return t
}
