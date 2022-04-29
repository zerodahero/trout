package handler

import (
	"log"
	"os"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

var api *slack.Client

// var botID string
var userID string

func InitApi(botToken, appToken string, debug bool) error {
	api = slack.New(
		botToken,
		slack.OptionDebug(debug),
		slack.OptionLog(log.New(os.Stdout, "api: ", log.Lshortfile|log.LstdFlags)),
		slack.OptionAppLevelToken(appToken),
	)

	var err error
	userID, _, err = getBotIDs(api)

	return err
}

func getBotIDs(api *slack.Client) (string, string, error) {
	// Get the bot's user ID
	auth, err := api.AuthTest()
	if err != nil {
		return "", "", err
	}

	return auth.UserID, auth.BotID, nil
}

func NewClient(debug bool) *socketmode.Client {
	return socketmode.New(
		api,
		socketmode.OptionDebug(debug),
		socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
	)
}

func GetUserInfo(userID string) (*slack.User, error) {
	return api.GetUserInfo(userID)
}
