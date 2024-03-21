package rotation

import "time"
import "fmt"
import "slices"
import "errors"

type User struct {
	Id   string
}

type Rotation struct {
	Users  []User
	Start  time.Time
	Period time.Duration
	persist bool
}

func New(persist bool) *Rotation {
	if persist {

	}
	loc, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		panic(err)
	}
	dur, err := time.ParseDuration("336h")
	if err != nil {
		panic(err)
	}
	return &Rotation{
		Users: []User{},
		Start: time.Date(2024, 03, 04, 8, 0, 0, 0, loc),
		Period: dur,
	}
}

func (r *Rotation) NextShift() []User {
	return r.Users[:3]
}

func (r *Rotation) Rotate() {
	r.Start = r.Start.Add(r.Period)
	r.Users = append(r.Users[3:], r.Users[:3]...)
}

func (r *Rotation) findUser(Id string) (int, bool) {
	for i, u := range r.Users {
		if u.Id == Id {
			return i, true
		}
	}
	return 0, false
}

func (r *Rotation) AddUser(Id string) error {
	if _, ok := r.findUser(Id); ok {
		return errors.New("already exists")
	}
	r.Users = append(r.Users, User{Id})
	fmt.Printf("\n%+v\n", r.Users)
	return nil
}

func (r *Rotation) NextUserShift(Id string) (time.Time, error) {
	i, ok := r.findUser(Id)
	if !ok {
		return time.Now(), errors.New("User not found")
	}
	durToShift := r.Period * time.Duration(i/3)
	return time.Now().Add(durToShift), nil
}

func (r *Rotation) PushBackShift(Id string) error {
	i, ok := r.findUser(Id)
	if !ok {
		return errors.New("User not found")
	}
	newSlot := (i / 3 + 1) * 3	
	if len(r.Users) < newSlot - 1 {
		return errors.New("Cannot move further back")
	}
	r.Users = slices.Delete(r.Users, i, i)
	r.Users = slices.Insert(r.Users, newSlot, User{Id})
	return nil
}

func readFile(path string) []User {
		f, err := os.Open(path)
		if err != nil {
			panic(err)
		}
 }
