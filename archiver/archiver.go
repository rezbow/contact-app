package archiver

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

type Status string

const (
	StatusInProgess Status = "in progress"
	StatusComplete  Status = "complete"
)

type ArchiveJob struct {
	done     chan struct{}
	result   string
	err      error
	progress atomic.Int32
	status   Status
	mu       sync.RWMutex
}

func (j *ArchiveJob) Run(ctx context.Context, sleepFunc func(time.Duration)) {
	defer close(j.done)
	for range 10 {
		select {
		case <-ctx.Done():
			j.mu.Lock()
			j.err = fmt.Errorf("Archive job failed: %w", ctx.Err())
			j.status = StatusComplete
			j.mu.Unlock()
			log.Printf("archive job failed: %q", ctx.Err().Error())
			return
		default:
			sleepFunc(time.Second * time.Duration(rand.Int31n(5)))
			log.Println("Job progess ", j.progress.Load())
			j.progress.Add(1)
		}
	}
	j.mu.Lock()
	j.result = "contacts.json"
	j.status = StatusComplete
	j.mu.Unlock()
	log.Println("job finished")
}

func (j *ArchiveJob) Done() <-chan struct{} {
	return j.done
}

func (j *ArchiveJob) Error() error {
	return j.err
}

func (j *ArchiveJob) Result() string {
	return j.result
}

func (j *ArchiveJob) Status() Status {
	return j.status
}

func (j *ArchiveJob) Progress() int {
	return int(j.progress.Load())
}

var archiver *Archiver

func init() {
	archiver = &Archiver{
		jobs:      make(map[string]*ArchiveJob),
		sleepFunc: time.Sleep,
	}
}

func GetArchiver() *Archiver {
	return archiver
}

type Archiver struct {
	mu        sync.Mutex
	jobs      map[string]*ArchiveJob
	sleepFunc func(time.Duration)
}

func (a *Archiver) Archive(ctx context.Context, userId string) *ArchiveJob {
	job := &ArchiveJob{
		done:   make(chan struct{}),
		status: StatusInProgess,
	}
	a.mu.Lock()
	a.jobs[userId] = job
	a.mu.Unlock()
	go job.Run(ctx, a.sleepFunc)
	return job
}

func (a *Archiver) GetJob(user_id string) *ArchiveJob {
	return a.jobs[user_id]
}
