package handler

import (
	"fmt"
	"time"

	"github.com/zerodahero/trout/database"
	"github.com/zerodahero/trout/parser"

	"github.com/slack-go/slack"
)

func BuildShoutTroutPasswordBlocks(attempt int, message string) []slack.Block {
	inputBlock := slack.NewInputBlock(
		fmt.Sprintf("shouttrout-%d", attempt),
		&slack.TextBlockObject{
			Type: slack.PlainTextType,
			Text: message,
		},
		slack.NewPlainTextInputBlockElement(
			&slack.TextBlockObject{
				Type: slack.PlainTextType,
				Text: "What's the super secret password?",
			},
			"",
		),
	)
	inputBlock.DispatchAction = true

	headerBlockMessage := "This will release *all* the baby shout trouts into the wild, starting right here IN THIS CHANNEL!\n\nAre you sure you wan to do that, RIGHT NOW?"
	if attempt >= 3 {
		headerBlockMessage = "Looks like you don't know the super secret password. ACCESS DENIED!"
	}

	headerBlock := slack.NewHeaderBlock(
		&slack.TextBlockObject{
			Type: slack.PlainTextType,
			Text: headerBlockMessage,
		},
	)

	blocks := []slack.Block{}
	blocks = append(blocks, headerBlock)
	if attempt < 3 {
		blocks = append(blocks, inputBlock)
	}

	return blocks
}

func HandleShoutTroutCommand(cmd slack.SlashCommand) (interface{}, error) {
	blocks := BuildShoutTroutPasswordBlocks(1, "Please enter the super secret password to continue.")

	return map[string]interface{}{"blocks": blocks}, nil
}

func HandleShoutTroutInteraction(a *slack.BlockAction, callback slack.InteractionCallback, attempt int) error {
	var blocks []slack.Block
	if a.Value != "Here, fishy fishy fishy!" {
		blocks = BuildShoutTroutPasswordBlocks(attempt+1, "Good try, but WRONG!")
		slack.PostWebhook(callback.ResponseURL, &slack.WebhookMessage{Blocks: &slack.Blocks{BlockSet: blocks}, ReplaceOriginal: true})
		return nil
	}

	blocks = []slack.Block{
		slack.NewHeaderBlock(
			&slack.TextBlockObject{
				Type: slack.PlainTextType,
				Text: "Access Granted!",
			},
		),
	}
	slack.PostWebhook(callback.ResponseURL, &slack.WebhookMessage{Blocks: &slack.Blocks{BlockSet: blocks}, ReplaceOriginal: true})

	err := releasePublicKudos(callback.Channel.ID, callback.User.ID)
	if err != nil {
		return err
	}

	return releasePrivateKudos(callback.Channel.ID, callback.User.ID)
}

func releasePublicKudos(channelID, userID string) error {
	kudos, err := database.GetUnsharedKudos(true)
	if err != nil {
		return err
	}
	err = notifyReleaseKudoCount(channelID, userID, true, len(kudos))
	if err != nil {
		return err
	}
	// No kudos, nothing to do
	if len(kudos) == 0 {
		return nil
	}

	shareTime := time.Now().UTC()
	prevUserID := ""
	threadTs := ""

	// Slack limits to ~1 post/s
	postLimiter := time.NewTicker(1 * time.Second).C
	// No clue what the limit is here, but we will limit anyway
	threadLimiter := time.NewTicker(350 * time.Millisecond).C
	for _, kudo := range kudos {
		if kudo.ToUserID != prevUserID {
			// start new thread
			<-postLimiter
			_, threadTs, err = api.PostMessage(channelID, slack.MsgOptionText(parser.WrapUserIdForMention(kudo.ToUserID), false))
			if err != nil {
				return err
			}
		}
		<-threadLimiter
		prevUserID = kudo.ToUserID
		_, _, err = api.PostMessage(channelID, slack.MsgOptionText(fmt.Sprintf("> %s\n - %s", kudo.Message, kudo.GetDisplayFrom(true)), false), slack.MsgOptionTS(threadTs))
		if err != nil {
			continue
		}

		err = kudo.MarkShared(shareTime)
		if err != nil {
			return err
		}
	}

	return nil
}

func releasePrivateKudos(channelID, userID string) error {
	kudos, err := database.GetUnsharedKudos(false)
	if err != nil {
		return err
	}
	err = notifyReleaseKudoCount(channelID, userID, false, len(kudos))
	if err != nil {
		return err
	}
	// No kudos, nothing to do
	if len(kudos) == 0 {
		return nil
	}

	limiter := time.NewTicker(1 * time.Second).C
	for _, kudo := range kudos {
		<-limiter
		_, _, err = api.PostMessage(kudo.ToUserID, slack.MsgOptionText(fmt.Sprintf("> %s\n - %s", kudo.Message, kudo.GetDisplayFrom(false)), false))
		if err != nil {
			continue
		}

		err = kudo.MarkShared(time.Now().UTC())
		if err != nil {
			return err
		}
	}

	return nil
}
