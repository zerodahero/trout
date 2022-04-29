package handler

import (
	"errors"
	"fmt"

	"github.com/zerodahero/trout/database"
	"github.com/zerodahero/trout/parser"

	"github.com/slack-go/slack"
)

func BuildCommandPayloadBlocks(kudo database.Kudo, message string) []slack.Block {
	var privateBlock, anonymousBlock slack.BlockElement

	if kudo.IsPublic {
		privateBlock = slack.NewButtonBlockElement(
			"",
			"private",
			&slack.TextBlockObject{
				Type: slack.PlainTextType,
				Text: "Make Private",
			},
		)
	} else {
		privateBlock = slack.NewButtonBlockElement(
			"",
			"public",
			&slack.TextBlockObject{
				Type: slack.PlainTextType,
				Text: "Make Public",
			},
		)
	}

	if kudo.IsAnonymous {
		anonymousBlock = slack.NewButtonBlockElement(
			"",
			"named",
			&slack.TextBlockObject{
				Type: slack.PlainTextType,
				Text: "Remove Anonymity",
			},
		)
	} else {
		anonymousBlock = slack.NewButtonBlockElement(
			"",
			"anonymous",
			&slack.TextBlockObject{
				Type: slack.PlainTextType,
				Text: "Make Anonymous",
			},
		)
	}

	return []slack.Block{
		slack.NewSectionBlock(
			&slack.TextBlockObject{
				Type:     slack.PlainTextType,
				Text:     "Shout out: " + kudo.Message,
				Verbatim: false,
			},
			nil,
			nil,
		),
		slack.NewSectionBlock(
			&slack.TextBlockObject{
				Type: slack.MarkdownType,
				Text: message,
			},
			nil,
			nil,
		),
		slack.NewActionBlock(
			fmt.Sprintf("kudo-%d", kudo.ID),
			privateBlock,
			anonymousBlock,
		),
	}
}

func HandleTroutCommand(cmd slack.SlashCommand, save func(*database.Kudo) error) (interface{}, error) {
	mentionCount := parser.GetMentionCount(cmd.Text)
	if mentionCount != 1 {
		err := notifyMissingToUser(cmd.ChannelID, cmd.UserID)
		if err != nil {
			fmt.Printf("failed posting message: %v", err)
		}
		return nil, err
	}

	kudo, err := database.NewKudoFromText(cmd.Text, cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse kudo: %v", err)
	}

	if kudo.FromUserID == kudo.ToUserID {
		notifySelfShoutOutNotAllowed(cmd.ChannelID, cmd.UserID)
		return nil, err
	}

	err = save(kudo)
	if err != nil {
		return nil, fmt.Errorf("failed to store kudo: %v", err)
	}

	blocks := BuildCommandPayloadBlocks(*kudo, "Thanks, got it!")

	return map[string]interface{}{"blocks": blocks}, nil
}

func HandleTroutInteraction(a *slack.BlockAction, callback slack.InteractionCallback, kudoID int) error {
	kudo, err := database.GetKudoByID(kudoID)
	if err != nil {
		return fmt.Errorf("could not find shout out: %v", err)
	}

	fmt.Println(kudo)

	switch a.Value {
	case "private":
		kudo.IsPublic = false
	case "public":
		kudo.IsPublic = true
	case "anonymous":
		kudo.IsAnonymous = true
	case "named":
		kudo.IsAnonymous = false
	default:
		return errors.New("unknown action value")
	}
	kudo.Save()

	blocks := BuildCommandPayloadBlocks(*kudo, "Successfully set shout out to be "+a.Value+"!")
	slack.PostWebhook(callback.ResponseURL, &slack.WebhookMessage{Blocks: &slack.Blocks{BlockSet: blocks}, ReplaceOriginal: true})

	return nil
}
