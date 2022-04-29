package parser

import (
	"errors"
	"regexp"
	"strings"
)

var whitespaceRegex = regexp.MustCompile(`\s{2,}`)
var userRegex = regexp.MustCompile(`<@([[:alnum:]]+)(\|[^>]+)?>`)

func GetMentionCount(text string) int {
	return strings.Count(text, "<@")
}

func ParseRecipientFromText(text string) (string, error) {
	to := userRegex.FindStringSubmatch(text)
	if to == nil {
		return "", errors.New("unable to find a user mention")
	}

	return to[1], nil
}

func ReplaceUserInText(text, userID, name string) string {
	userIDRegex := regexp.MustCompile(`<@` + userID + `(\|[^>]+)?>`)
	return userIDRegex.ReplaceAllString(text, name)
}

func RemoveWhitespace(text string) string {
	text = whitespaceRegex.ReplaceAllString(text, "")
	text = strings.TrimSpace(text)

	return text
}

func WrapUserIdForMention(userID string) string {
	return `<@` + userID + `>`
}
