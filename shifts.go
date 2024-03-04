package main

import "time"
import "fmt"
import "slices"
import "errors"

type user struct {
	id   string
}

type Rotation struct {
	users  []user
	start  time.Time
	period time.Duration
}

func New() *Rotation {
	loc, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		panic(err)
	}
	dur, err := time.ParseDuration("336h")
	if err != nil {
		panic(err)
	}
	return &Rotation{
		users: []user{},
		start: time.Date(2024, 03, 04, 8, 0, 0, 0, loc),
		period: dur,
	}
}

func (r *Rotation) NextShift() []user {
	return r.users[:3]
}

func (r *Rotation) Rotate() {
	r.start = r.start.Add(r.period)
	r.users = append(r.users[3:], r.users[:3]...)
}

func (r *Rotation) findUser(id string) (int, bool) {
	for i, u := range r.users {
		if u.id == id {
			return i, true
		}
	}
	return 0, false
}

func (r *Rotation) AddUser(id string) error {
	if _, ok := r.findUser(id); ok {
		return errors.New("already exists")
	}
	r.users = append(r.users, user{id})
	fmt.Printf("\n%+v\n", r.users)
	return nil
}

func (r *Rotation) NextUserShift(id string) (time.Time, error) {
	i, ok := r.findUser(id)
	if !ok {
		return time.Now(), errors.New("user not found")
	}
	durToShift := r.period * time.Duration(i/3)
	return time.Now().Add(durToShift), nil
}

func (r *Rotation) PushBackShift(id string) error {
	i, ok := r.findUser(id)
	if !ok {
		return errors.New("User not found")
	}
	newSlot := (i / 3 + 1) * 3	
	if len(r.users) < newSlot - 1 {
		return errors.New("Cannot move further back")
	}
	r.users = slices.Delete(r.users, i, i)
	r.users = slices.Insert(r.users, newSlot, user{id})
	return nil
}
