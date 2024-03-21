package controller

import (
	"github.com/lexysoda/officeduty/slack"
	"github.com/lexysoda/officeduty/rotation"
)

type Controller struct {
	s *slack.Slack
	r *rotation.Rotation
}

func New(s *slack.Slack, r *rotation.Rotation) *Controller{
	return &Controller{s, r}
}

func (c *Controller) Start() {
	c.s.Start(c.handleSlash, c.handleMention)
}

func (c *Controller) handleSlash(id string) {
	msg := "You are currently not in the rotation"
	if date, err := c.r.NextUserShift("<@" + id + ">"); err == nil {
		msg = "Your next shift: " + date.Format("Jan 02")
	}
	c.s.SendEphemeral(msg, id)
}

func (c *Controller) handleMention(args ...string) {
	if len(args) != 3 || args[1] != "add" {
		c.sendShift()
		return
	}
	err := c.r.AddUser(args[2])
	if err != nil {
		c.s.SendMessage(args[2] + " is already on the list")
	} else {
		c.s.SendMessage(args[2] + " added to the list")
	}
}

func (c *Controller) handleCommand(args ...string) {
	switch {
	case len(args) == 0:
		return
	case args[0] == "/officeshift":
		return
	default:
		c.sendShift()
	}
}

func (c *Controller) sendShift() {
	start := c.r.Start.Format("02.01.")
	end := c.r.Start.Add(c.r.Period).Format("02.01.")
	users := c.r.NextShift()
	ids := []string{}
	for i, u := range users {
		ids[i] = u.Id
	}

	c.s.SendShift(start, end, ids)
}
