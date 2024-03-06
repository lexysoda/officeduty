package main

type Controller struct {
	s *Slack
	r *Rotation
}

func (c *Controller) handleSlash(args ...string) {
	if len(args) != 2 {
		return
	}
	id := args[2]
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
	start := c.r.start.Format("02.01.")
	end := c.r.start.Add(c.r.period).Format("02.01.")
	users := c.r.NextShift()
	ids := []string{}
	for i, u := range users {
		ids[i] = u.id
	}

	c.s.SendShift(start, end, ids)
}
