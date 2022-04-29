package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/zerodahero/trout/database"
	"github.com/zerodahero/trout/handler"
	"github.com/zerodahero/trout/parser"

	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

var debug bool

func init() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Print("Error loading .env file")
	}

	debugString := os.Getenv("DEBUG")
	debug, err = strconv.ParseBool(debugString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "DEBUG not set or not parseable, defaulting to false.\n")
		debug = false
	}

	appToken := os.Getenv("SLACK_APP_TOKEN")
	if appToken == "" {
		fmt.Fprintf(os.Stderr, "SLACK_APP_TOKEN must be set.\n")
		os.Exit(1)
	}

	if !strings.HasPrefix(appToken, "xapp-") {
		fmt.Fprintf(os.Stderr, "SLACK_APP_TOKEN must have the prefix \"xapp-\".")
	}

	botToken := os.Getenv("SLACK_BOT_TOKEN")
	if botToken == "" {
		fmt.Fprintf(os.Stderr, "SLACK_BOT_TOKEN must be set.\n")
		os.Exit(1)
	}

	if !strings.HasPrefix(botToken, "xoxb-") {
		fmt.Fprintf(os.Stderr, "SLACK_BOT_TOKEN must have the prefix \"xoxb-\".")
	}

	err = database.InitDB("./trout.db")
	if err != nil {
		log.Fatal(err)
	}

	err = handler.InitApi(botToken, appToken, debug)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {

	client := handler.NewClient(debug)

	go func() {
		for evt := range client.Events {
			switch evt.Type {
			case socketmode.EventTypeConnecting:
				fmt.Println("Connecting to Slack with Socket Mode...")
			case socketmode.EventTypeConnectionError:
				fmt.Println("Connection failed. Retrying later...")
			case socketmode.EventTypeConnected:
				fmt.Println("Connected to Slack with Socket Mode.")
			case socketmode.EventTypeEventsAPI:
				eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
				if !ok {
					fmt.Printf("Ignored %+v\n", evt)

					continue
				}

				fmt.Printf("Event received: %+v\n", eventsAPIEvent)

				client.Ack(*evt.Request)

				switch eventsAPIEvent.Type {
				case slackevents.CallbackEvent:
					innerEvent := eventsAPIEvent.InnerEvent
					switch ev := innerEvent.Data.(type) {
					case *slackevents.AppMentionEvent:
						handler.HandleMention(ev, saveKudoWithUser)
					case *slackevents.MemberJoinedChannelEvent:
						fmt.Printf("user %q joined to channel %q", ev.User, ev.Channel)
					}
				default:
					client.Debugf("unsupported Events API event received")
				}
			case socketmode.EventTypeInteractive:
				callback, ok := evt.Data.(slack.InteractionCallback)
				if !ok {
					fmt.Printf("Ignored %+v\n", evt)

					continue
				}

				fmt.Printf("Interaction received: %+v\n", callback)

				var payload interface{}

				switch callback.Type {
				case slack.InteractionTypeBlockActions:
					// See https://api.slack.com/apis/connections/socket-implement#button
					for _, a := range callback.ActionCallback.BlockActions {
						var err error
						actionType := strings.Split(a.BlockID, "-")
						switch actionType[0] {
						case "kudo":
							kudoID, _ := strconv.Atoi(actionType[1])
							err = handler.HandleTroutInteraction(a, callback, kudoID)
						case "shouttrout":
							attempt, _ := strconv.Atoi(actionType[1])
							err = handler.HandleShoutTroutInteraction(a, callback, attempt)
						}
						if err != nil {
							fmt.Printf("Error handling interaction: %v", err)
						}
					}
					client.Debugf("button clicked!")
				case slack.InteractionTypeShortcut:
				case slack.InteractionTypeViewSubmission:
					// See https://api.slack.com/apis/connections/socket-implement#modal
				case slack.InteractionTypeDialogSubmission:
				default:

				}

				client.Ack(*evt.Request, payload)
			case socketmode.EventTypeSlashCommand:
				cmd, ok := evt.Data.(slack.SlashCommand)
				if !ok {
					fmt.Printf("Ignored %+v\n", evt)

					continue
				}

				client.Debugf("Slash command received: %+v", cmd)

				var payload interface{}
				var err error

				switch cmd.Command {
				case "/trout":
					payload, err = handler.HandleTroutCommand(cmd, saveKudoWithUser)
				case "/shout-trout":
					payload, err = handler.HandleShoutTroutCommand(cmd)
				default:
					fmt.Fprintf(os.Stderr, "Unexpected slash command received: %s\n", cmd.Command)
				}

				if err != nil {
					fmt.Printf("Error handling slash command: %v", err)
				}

				client.Ack(*evt.Request, payload)
			default:
				fmt.Fprintf(os.Stderr, "Unexpected event type received: %s\n", evt.Type)
			}
		}
	}()

	client.Run()
}

func saveKudoWithUser(kudo *database.Kudo) error {
	user, err := database.GetOrFetchUser(kudo.ToUserID, handler.GetUserInfo)
	if err != nil {
		return fmt.Errorf("failed to get user info: %v", err)
	}

	// Get the "from" user as well to make sure they're in the DB
	_, err = database.GetOrFetchUser(kudo.FromUserID, handler.GetUserInfo)
	if err != nil {
		return fmt.Errorf("failed to get user info: %v", err)
	}

	kudo.Message = parser.ReplaceUserInText(kudo.Message, user.SlackID, user.DisplayName)

	err = kudo.Save()
	if err != nil {
		return fmt.Errorf("failed to save kudo: %v", err)
	}

	return nil
}
