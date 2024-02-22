
package main

import (
   "fmt"
   "log"
   "os"
   "reflect"

   "github.com/slack-go/slack/socketmode"
   "github.com/slack-go/slack"
   "github.com/slack-go/slack/slackevents"
)

type Slack struct {
	slackClient *slack.Client
	socketClient *socketmode.Client
	r *Rotation
}

func NewSlack(debug bool, r *Rotation) *Slack {
   appToken := os.gentenv
   botToken := os.getenv

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

   return &Slack{api, client, r}
}


func (s *Slack) Start() {
   socketmodeHandler := socketmode.NewSocketmodeHandler(s.socketClient)

   socketmodeHandler.HandleEvents(slackevents.AppMention, handleMention)
   socketmodeHandler.HandleSlashCommand("/officeshift", handleSlash)
   socketmodeHandler.HandleDefault(handleEvents)

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

func handleMention(evt *socketmode.Event, client *socketmode.Client) {
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

   fmt.Printf("We have been mentionned in %v\n", ev.Channel)
   _, _, err := client.Client.PostMessage(ev.Channel, slack.MsgOptionText("Yes, hello. Here is the current shift:", false))
	if err != nil {
		fmt.Printf("failed posting message: %v", err)
	}
}

func handleSlash(evt *socketmode.Event, client *socketmode.Client) {
   cmd, ok := evt.Data.(slack.SlashCommand)
	if !ok {
		fmt.Printf("Ignored %+v\n", evt)
		return
	}

	client.Debugf("Slash command received: %+v", cmd)

	payload := map[string]interface{}{
		"blocks": []slack.Block{
			slack.NewSectionBlock(
				&slack.TextBlockObject{
					Type: slack.MarkdownType,
					Text: "foo",
				},
				nil,
				slack.NewAccessory(
					slack.NewButtonBlockElement(
						"",
						"somevalue",
						&slack.TextBlockObject{
							Type: slack.PlainTextType,
							Text: "bar",
						},
					),
				),
			),
		}}
	client.Ack(*evt.Request, payload)
}
