package rotation

import (
	"errors"
	"regexp"

	"github.com/slack-go/slack"
)

var regex = regexp.MustCompile(`<@((?:U|W)[A-Z0-9]{1,21})(?:\|.*)?>`)

type User string

func (u User) ToBlockElement() *slack.RichTextSectionUserElement {
	return slack.NewRichTextSectionUserElement(
		string(u),
		nil,
	)
}

func FromUnescaped(s string) (User, error) {
	matches := regex.FindStringSubmatch(s)
	if len(matches) != 2 {
		return "", errors.New("Failed to parse user")
	}
	return User(matches[1]), nil
}
