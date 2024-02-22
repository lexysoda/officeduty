package main

import "time"
import "errors"

type user struct {
	id	string
	name	string
}

type Rotation struct {
	users []user
	start time.Time
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
		start: time.Date(2024, 02, 19, 8, 0, 0, 0, loc),
		period: dur,
	}
}

func (r *Rotation) NextShift() []user{
	return r.users[:3]
}

func (r *Rotation) Rotate() {
	r.users = append(r.users[3:], r.users[:3]...)
}

func (r *Rotation) findUser(id string) (int, bool) {
	for i, u := range(r.users) {
		if u.id == id {
			return i, true
		}
	}
	return 0, false
}

func (r *Rotation) AddUser(id, name string) error {
	if _, ok := r.findUser(id); ok {
		return errors.New("already exists")
	}
	r.users = append(r.users, user{id, name})
	return nil
}

func (r *Rotation) NextUserShift(id string) (time.Time, error) {
	i, ok := r.findUser(id)
	if !ok {
		return time.Now(), errors.New("user not found")
	}
	durToShift := r.period * time.Duration(i / 3)
	return time.Now().Add(durToShift), nil
}
