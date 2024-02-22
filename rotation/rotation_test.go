package rotation

import "testing"
import "time"
import "github.com/lexysoda/officeduty/user"

func TestRotation(t *testing.T) {
	r := &Rotation{
		Users: []user.User{},
		start: time.Now(),
		interval: time.Duration(10 * time.Second),
	}

	r.AddUser(user.User("user1"))
	r.AddUser(user.User("user2"))
	r.AddUser(user.User("user3"))
	r.AddUser(user.User("user4"))
	r.AddUser(user.User("user5"))
	r.AddUser(user.User("user6"))
	r.AddUser(user.User("user7"))
	r.AddUser(user.User("user8"))
	
	t.Logf("%+v\n", r.NextShift())

	nextShift, err := r.NextUserShift(user.User("user1"))
	if err != nil {
		t.Fatalf("kek")
	}
	t.Logf("%+v\n", nextShift)
	t.Logf("%+v\n", r.NextShift())

	r.rotate()
	nextShift, err = r.NextUserShift(user.User("user1"))
	if err != nil {
		t.Fatalf("kek")
	}
	t.Logf("%+v\n", nextShift)
	t.Logf("%+v\n", r.NextShift())

	r.Start()
	timer := time.NewTicker(time.Second * 15)
	<-timer.C
	r.Stop()
	t.Logf("%+v\n", r.NextShift())
}
