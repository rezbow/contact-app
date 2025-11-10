package archiver

import (
	"context"
	"errors"
	"testing"
	"time"
)

type MockSleeper struct {
	sleepCalls int
}

func (m *MockSleeper) Sleep(d time.Duration) {
	m.sleepCalls++
}

// archiver archives the contact
// database into filesystem and it
// gives us the path in filesystem
func TestArchiver(t *testing.T) {
	t.Run("succesful archive job", func(t *testing.T) {
		sleeper := &MockSleeper{}
		archiver := &Archiver{
			jobs:      make(map[string]*ArchiveJob),
			sleepFunc: sleeper.Sleep,
		}
		job := archiver.Archive(context.Background(), "user_id")
		want := "contacts.json"
		// wait for job to be done
		<-job.Done()

		if job.Error() != nil {
			t.Errorf("job resulted in err:%v, exptected none", job.Error())
		}
		if job.Result() != want {
			t.Errorf("got result %q, wanted %q", job.Result(), want)
		}
		if job.Status() != StatusComplete {
			t.Errorf("got status %q, wanted %q", job.Status(), StatusComplete)
		}
		if sleeper.sleepCalls != 10 {
			t.Errorf("got %d calls to sleep, wanted %d", sleeper.sleepCalls, 10)
		}
	})

	t.Run("cancel archive job results in job error", func(t *testing.T) {
		sleeper := &MockSleeper{}
		archiver := &Archiver{
			jobs:      make(map[string]*ArchiveJob),
			sleepFunc: sleeper.Sleep,
		}
		ctx, cancel := context.WithCancel(context.Background())
		job := archiver.Archive(ctx, "user_id")
		cancel()
		<-job.Done()

		if job.Error() == nil {
			t.Errorf("job resulted in no error, we wanted one")
		}
		if errors.Unwrap(job.Error()) != context.Canceled {
			t.Errorf("wanted %v as error, but it isnt", context.Canceled)
		}
		if job.Status() != StatusComplete {
			t.Errorf("got status %q, wanted %q", job.Status(), StatusComplete)
		}
		if job.Result() != "" {
			t.Errorf("got result %q, wanted '' ", job.Result())
		}
	})
}
