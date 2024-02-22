package main

import (
	"github.com/lexysoda/officeduty/controller"
	"github.com/lexysoda/officeduty/rotation"
)

func main() {
	r := rotation.New()
	r.AddUser(rotation.User("1"))
	r.AddUser(rotation.User("2"))
	r.AddUser(rotation.User("+1"))
	r.AddUser(rotation.User("+2"))
	r.AddUser(rotation.User("UB048064V"))

	c := controller.New(r, true)

	c.Start()
}
