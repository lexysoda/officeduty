package main

func main() {
	r := New()
	r.users = []user{
		user{"1"},
		user{"2"},
		user{"@roman"},
		user{"+1"},
		user{"+2"},
	}
	s := NewSlack(true,"C06LSFGJ0HE", r)
	s.Start()
}
