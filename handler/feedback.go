package handler

import (
	"fmt"

	"github.com/slack-go/slack"
)

func notifyMissingToUser(channelID, userID string) error {
	return notifyUser(channelID, userID, "Hmmm, who's this about? Please try again and tag the single user you want to shout out.")
}

func notifySelfShoutOutNotAllowed(channelID, userID string) error {
	return notifyUser(channelID, userID, "Glad to hear you're doing some great work, but I don't do self shout-outs.")
}

func notifyMultipleUserNotSupported(channelID, userID string) error {
	return notifyUser(channelID, userID, "Sorry, shouting out multiple users at once is not supported (yet)? Please try again and tag each user in individual messages.")
}

func notifyKudoReceived(channelID, userID string) error {
	return notifyUser(channelID, userID, "Got it! You're awesome, thanks!")
}

func notifyReleaseKudoCount(channelID, userID string, public bool, count int) error {
	var visibility string
	if public {
		visibility = "public"
	} else {
		visibility = "private"
	}

	message := fmt.Sprintf("Releasing %d %s shout outs!", count, visibility)

	return notifyUser(channelID, userID, message)
}

func notifyUser(channelID, userID, message string) error {
	_, err := api.PostEphemeral(channelID, userID, slack.MsgOptionText(message, false))
	return err
}
