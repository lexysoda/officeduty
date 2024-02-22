package rotation

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"time"
)

type Rotation struct {
	users    []User
	start    time.Time
	interval time.Duration
	stop     context.CancelFunc
	notify   chan struct{}
}

type Shift struct {
	Users []User
	Start time.Time
	End   time.Time
}

func New() *Rotation {
	loc, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		panic(err)
	}
	dur, err := time.ParseDuration("30s")
	if err != nil {
		panic(err)
	}
	return &Rotation{
		users:    []User{},
		start:    time.Date(2024, 04, 04, 8, 0, 0, 0, loc),
		interval: dur,
	}
}

func (r *Rotation) NextShift() Shift {
	return Shift{
		r.users[:3],
		r.start,
		r.start.Add(r.interval),
	}
}

func (r *Rotation) rotate() {
	r.start = r.start.Add(r.interval)
	r.users = append(r.users[3:], r.users[:3]...)
	slog.Debug("Rotated", "users", r.users)
}

func (r *Rotation) AddUser(user User) error {
	if slices.Contains(r.users, user) {
		return errors.New("already exists")
	}
	r.users = append(r.users, user)
	return nil
}

func (r *Rotation) NextUserShift(user User) (time.Time, error) {
	i := slices.Index(r.users, user)
	if i == -1 {
		slog.Debug("User not found", "user", user)
		return time.Now(), errors.New("User not found")
	}
	durToShift := r.interval * time.Duration(i/3)
	return time.Now().Add(durToShift), nil
}

func (r *Rotation) PushBackShift(u User) error {
	i := slices.Index(r.users, u)
	if i == -1 {
		slog.Debug("User not found", "user", u)
		return errors.New("User not found")
	}
	newSlot := (i/3 + 1) * 3
	if len(r.users)-1 < newSlot {
		slog.Debug("Cannot move further back", "user", u)
		return errors.New("Cannot move further back")
	}
	usersNew := append([]User{}, r.users[:i]...)
	usersNew = append(usersNew, r.users[i+1:newSlot+1]...)
	usersNew = append(usersNew, r.users[i])
	usersNew = append(usersNew, r.users[newSlot+1:]...)
	r.users = usersNew
	slog.Debug("Pushed back", "user", u, "users", r.users)
	if i < 3 {
		r.notify <- struct{}{}
	}
	return nil
}

func (r *Rotation) nextShift() time.Time {
	next := r.start
	for next.Before(time.Now()) {
		next = next.Add(r.interval)
	}
	return next
}

func (r *Rotation) setTimer(ctx context.Context) {
	dur := r.nextShift().Sub(time.Now())
	t := time.NewTimer(dur)
	slog.Debug("Rotation timer set", "duration", dur)
	go func() {
		select {
		case <-t.C:
			r.rotate()
			r.notify <- struct{}{}
			r.setTimer(ctx)
		case <-ctx.Done():
			slog.Info("Rotation timer stopped")
			t.Stop()
			return
		}
	}()
}

func (r *Rotation) Start() chan struct{} {
	ctx, cancel := context.WithCancel(context.Background())
	r.setTimer(ctx)
	r.stop = cancel
	r.notify = make(chan struct{})
	return r.notify
}

func (r *Rotation) Stop() {
	r.stop()
}
