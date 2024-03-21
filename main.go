package main

import (
	"github.com/lexysoda/officeduty/slack"
	"github.com/lexysoda/officeduty/rotation"
	"github.com/lexysoda/officeduty/controller"
)

func main() {
	r := rotation.New()
	r.Users = []rotation.User{
		rotation.User{"1"},
		rotation.User{"2"},
		rotation.User{"@roman"},
		rotation.User{"+1"},
		rotation.User{"+2"},
	}
	s := slack.New(true,"C06LSFGJ0HE")
	c := controller.New(s, r)

	c.Start()
}
