package main

import (
   "fmt"
)

func main() {
	r := New()
	s := NewSlack(true, r)
	s.Start()
}
