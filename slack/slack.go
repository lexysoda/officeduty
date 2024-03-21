package slack

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

type Slack struct {
	slackClient  *slack.Client
	socketClient *socketmode.Client
	channel      string
}

func New(debug bool, channel string) *Slack {
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

	return &Slack{api, client, channel}
}

func (s *Slack) Start(slash func(string), mention func(...string)) {
	socketmodeHandler := socketmode.NewSocketmodeHandler(s.socketClient)
	socketmodeHandler.HandleEvents(slackevents.AppMention, s.handleMention(mention))
	socketmodeHandler.HandleSlashCommand("/officeshift", s.handleSlash(slash))

	//t := time.NewTicker(s.r.period)
	//go func() {
	//	for {
	//		<-t.C
	//		s.r.Rotate()
	//		s.sendShift("C06LSFGJ0HE")
	//	}
	//}()

	socketmodeHandler.RunEventLoop()
}

func (s *Slack) handleMention(callback func (...string)) func(*socketmode.Event, *socketmode.Client) {
	return func(evt *socketmode.Event, client *socketmode.Client) {
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

		callback(strings.Fields(ev.Text)...)
	}
}

func (s *Slack) handleSlash(callback func(string)) func(*socketmode.Event, *socketmode.Client) {
	return func(evt *socketmode.Event, client *socketmode.Client) {
		cmd, ok := evt.Data.(slack.SlashCommand)
		if !ok {
			fmt.Printf("Ignored %+v\n", evt)
			return
		}

		client.Debugf("Slash command received: %+v", cmd)

		client.Ack(*evt.Request)
		callback(cmd.UserID)
	}
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

	s.SendMessage(fmt.Sprintf(
		msg,
		start,
		end,
		users[0],
		users[1],
		users[2],
	))
}

