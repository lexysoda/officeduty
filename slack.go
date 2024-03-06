package main

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

type Slack struct {
	slackClient  *slack.Client
	socketClient *socketmode.Client
	channel      string
	r            *Rotation
}

func NewSlack(debug bool, channel string, r *Rotation) *Slack {
	appToken := os.Getenv("APP_TOKEN")
	botToken := os.Getenv("BOT_TOKEN")

	api := slack.New(
		botToken,
		slack.OptionDebug(debug),
		slack.OptionLog(log.New(os.Stdout, "api: ", log.Lshortfile|log.LstdFlags)),
		slack.OptionAppLevelToken(appToken),
	)

	client := socketmode.New(
		api,
		socketmode.OptionDebug(debug),
		socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
	)

	return &Slack{api, client, channel, r}
}

func (s *Slack) Start() {
	socketmodeHandler := socketmode.NewSocketmodeHandler(s.socketClient)

	socketmodeHandler.HandleEvents(slackevents.AppMention, s.handleMention)
	socketmodeHandler.HandleSlashCommand("/officeshift", s.handleSlash)
	socketmodeHandler.HandleDefault(handleEvents)

	t := time.NewTicker(s.r.period)
	go func() {
		for {
			<-t.C
			s.r.Rotate()
			s.sendShift("C06LSFGJ0HE")
		}
	}()

	socketmodeHandler.RunEventLoop()
}

func handleEvents(evt *socketmode.Event, client *socketmode.Client) {
	fmt.Println("middlewareEventsAPI")

	fmt.Println(evt.Type)

	if evt.Type == socketmode.EventTypeErrorBadMessage {
		fmt.Println(reflect.TypeOf(evt.Data))
		if inner, ok := evt.Data.(*socketmode.ErrorBadMessage); ok {
			fmt.Println("HELLLLOOO")
			fmt.Printf("%+v\n", inner)
		}
	}

	eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
	if !ok {
		fmt.Printf("Ignored %+v\n", evt)
		return
	}

	fmt.Printf("Event received: %+v\n", eventsAPIEvent)

	client.Ack(*evt.Request)
}

func (s *Slack) handleMention(evt *socketmode.Event, client *socketmode.Client) {
	eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
	if !ok {
		fmt.Printf("Ignored %+v\n", evt)
		return
	}

	client.Ack(*evt.Request)

	ev, ok := eventsAPIEvent.InnerEvent.Data.(*slackevents.AppMentionEvent)
	if !ok {
		fmt.Printf("Ignored %+v\n", ev)
		return
	}

	args := strings.Fields(ev.Text)
	if len(args) == 3 && args[1] == "add" {
		err := s.r.AddUser(args[2])
		if err != nil {
			s.sendMessage(fmt.Sprintf("%s is already on the list", args[2]), ev.Channel)
			return
		}
		return
	}
	s.sendShift(ev.Channel)
}

func (s *Slack) handleSlash(evt *socketmode.Event, client *socketmode.Client) {
	cmd, ok := evt.Data.(slack.SlashCommand)
	if !ok {
		fmt.Printf("Ignored %+v\n", evt)
		return
	}

	client.Debugf("Slash command received: %+v", cmd)

	payload := "Error"
	switch cmd.Command {
	case "/officeshift":
		date, err := s.r.NextUserShift("<@" + cmd.UserID + ">")
		if err != nil {
			payload = "You are currently not in the rotation"
			break
		}
		payload = "Your next shift: " + date.Format("Jan 02")
	default:
	}

	block := slack.NewTextBlockObject("mrkdwn", payload, false, false)
	client.Ack(*evt.Request, {})
}

func (s *Slack) SendEphemeral(m string, user string) {
	_, err := s.socketClient.Client.PostEphemeral(s.channel, user, slack.MsgOptionText(m, false))
	if err != nil {
		fmt.Printf("failed posting message: %v", err)
	}
}

func (s *Slack) SendMessage(m string) {
	_, _, err := s.socketClient.Client.PostMessage(s.channel, slack.MsgOptionText(m, false))
	if err != nil {
		fmt.Printf("failed posting message: %v", err)
	}
}

func (s *Slack) SendShift(start, end string, users []string) {
	msg :=
		`🚨*New Office Duty Shift*🚨
Start: %s    End: %s
--------------------------------
- %s
- %s
- %s
--------------------------------

<https://stroeerdigitalgroup.atlassian.net/wiki/spaces/MBR/pages/3528261726/Labs+Berlin+Permanent+Office+Management+and+Maintenance|Click here for more information>

- /officeshift to see your next shift and the option to reschedule
- /pushbackshift to move your shift to the next rotation
- /nextshift to see your next shift
- /fullshift to see the full list of upcoming shifts
`

	s.sendMessage(fmt.Sprintf(
		msg,
		start,
		end,
		users[0].id,
		users[1].id,
		users[2].id,
	))
}

func (s *Slack) sendFullRotation(channel string) {
	var sb strings.Builder
	for i, u := range s.r.users {
		if i%3 == 0 {
			fmt.Fprintf(&sb, "------\n")
		}
		fmt.Fprintf(&sb, "%s\n", u.id)
	}
	s.sendMessage(sb.String(), channel)
}
