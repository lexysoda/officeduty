package rotation

import "testing"
import "time"

func TestRotation(t *testing.T) {
	r := &Rotation{
		users: []user{},
		start: time.Now(),
		period: time.Duration(24 * time.Hour),
	}

	r.AddUser("123", "kek")
	r.AddUser("123", "kek")
	r.AddUser("1234", "kkkek")
	r.AddUser("2345", "knecht")
	r.AddUser("043", "foo")
	
	t.Logf("%+v\n", r.users)
	t.Logf("%+v\n", r.NextShift())


	nextShift, err := r.NextUserShift("043")
	if err != nil {
		t.Fatalf("kek")
	}
	t.Logf("%+v\n", nextShift)

	r.Rotate()
	nextShift, err = r.NextUserShift("043")
	if err != nil {
		t.Fatalf("kek")
	}
	t.Logf("%+v\n", nextShift)
}
