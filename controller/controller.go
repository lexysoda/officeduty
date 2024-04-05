package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/lexysoda/officeduty/rotation"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

type Controller struct {
	s       *slack.Client
	r       *rotation.Rotation
	secret  string
	channel string
	*http.ServeMux
}

func New(r *rotation.Rotation, debug bool) *Controller {
	if debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	botToken := os.Getenv("BOT_TOKEN")
	signingSecret := os.Getenv("SIGNING_SECRET")
	slog.Debug(fmt.Sprintf("botToken: %s", botToken))
	slog.Debug(fmt.Sprintf("signingSecret: %s", signingSecret))

	s := slack.New(
		botToken,
		slack.OptionDebug(false),
		slack.OptionLog(log.New(os.Stderr, "slack-go/slack: ", log.Lshortfile|log.LstdFlags)),
	)

	return &Controller{s, r, signingSecret, "C06LSFGJ0HE", http.NewServeMux()}
}

func (c *Controller) validate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			slog.Error("Failed to read body", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		sv, err := slack.NewSecretsVerifier(r.Header, c.secret)
		if err != nil {
			slog.Error("Failed to create secrets verifier", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if _, err := sv.Write(body); err != nil {
			slog.Error("Failed to write hmac", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err := sv.Ensure(); err != nil {
			slog.Error("Failed to validate", "error", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		r.Body = io.NopCloser(strings.NewReader(string(body)))
		next(w, r)
	}
}

func (c *Controller) Start() {
	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		if err := c.validate(w, r); err != nil {
			slog.Error("Failed to validate", "error", err)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			slog.Error("Failed to read body", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if eventsAPIEvent.Type == slackevents.URLVerification {
			var r *slackevents.ChallengeResponse
			err := json.Unmarshal([]byte(body), &r)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text")
			w.Write([]byte(r.Challenge))
		}

		if eventsAPIEvent.Type == slackevents.CallbackEvent {
			innerEvent := eventsAPIEvent.InnerEvent
			switch ev := innerEvent.Data.(type) {
			case *slackevents.AppMentionEvent:
				c.handleMention(strings.Fields(ev.Text)...)
			}
		}
	})

	http.HandleFunc("/slash/officeshift", func(w http.ResponseWriter, r *http.Request) {
		if err := c.validate(w, r); err != nil {
			slog.Error("Failed to validate", "error", err)
			return
		}

		s, err := slack.SlashCommandParse(r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		c.sendUserShift(rotation.User(s.UserID))
	})

	http.HandleFunc("/interaction", func(w http.ResponseWriter, r *http.Request) {
		if err := c.validate(w, r); err != nil {
			slog.Error("Failed to validate", "error", err)
			return
		}

		var i slack.InteractionCallback
		if err := json.Unmarshal([]byte(r.FormValue("payload")), &i); err != nil {
			slog.Error("Failed to unmarshal", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		action := i.ActionCallback.BlockActions[0].ActionID
		if action != "reschedule" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		u := rotation.User(i.User.ID)
		if err := c.r.PushBackShift(u); err != nil {
			c.sendFailedReschedule(u, i.ResponseURL)
			return
		}
		c.sendUserShiftReplace(u, i.ResponseURL)
	})

	notify := c.r.Start()

	go func() {
		for {
			<-notify
			c.sendShiftPlan()
		}
	}()

	slog.Info("Server listening", "port", 1337)
	http.ListenAndServe(":1337", nil)
}

func (c *Controller) handleMention(args ...string) {
	if len(args) != 3 || args[1] != "add" {
		c.sendShiftPlan()
		return
	}
	u, err := rotation.FromUnescaped(args[2])
	if err != nil {
		c.sendTextMessage("Failed to parse user")
	}
	err = c.r.AddUser(u)
	if err != nil {
		c.sendTextMessage(args[2] + " is already on the list")
	} else {
		c.sendTextMessage(args[2] + " added to the list")
	}
}
