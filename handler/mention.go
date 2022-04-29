package handler

import (
	"fmt"

	"github.com/zerodahero/trout/database"
	"github.com/zerodahero/trout/parser"

	"github.com/slack-go/slack/slackevents"
)

func HandleMention(ev *slackevents.AppMentionEvent, save func(*database.Kudo) error) {
	mentionCount := parser.GetMentionCount(ev.Text)
	if mentionCount < 2 {
		err := notifyMissingToUser(ev.Channel, ev.User)
		if err != nil {
			fmt.Printf("failed posting message: %v", err)
		}
		return
	}
	if mentionCount > 2 {
		err := notifyMultipleUserNotSupported(ev.Channel, ev.User)
		if err != nil {
			fmt.Printf("failed posting message: %v", err)
		}
		return
	}

	kudo, err := database.NewKudoFromMentionEvent(ev, userID)
	if err != nil {
		fmt.Printf("failed to parse kudo: %v", err)
		return
	}

	if kudo.FromUserID == kudo.ToUserID {
		err := notifySelfShoutOutNotAllowed(ev.Channel, ev.User)
		if err != nil {
			fmt.Printf("failed posting message: %v", err)
		}
		return
	}

	err = save(kudo)
	if err != nil {
		fmt.Printf("failed to store kudo: %v", err)
		return
	}

	err = notifyKudoReceived(ev.Channel, ev.User)
	if err != nil {
		fmt.Printf("failed posting acknowledgement: %v", err)
	}
}
