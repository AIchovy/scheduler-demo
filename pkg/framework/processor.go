package framework

import (
	"time"

	"github.com/avast/retry-go"

	job2 "github.com/mosesyou/scheduler-demo/pkg/job"
	"github.com/mosesyou/scheduler-demo/pkg/store"
)

// Processor
type Processor struct {
	store     store.JobStore
	workQueue *workQueue
}

func NewProcessor(store store.JobStore, queue *workQueue) *Processor {
	return &Processor{
		store:     store,
		workQueue: queue,
	}
}

func (p *Processor) Run(workers int, stopCh <-chan struct{}) {
	defer p.workQueue.ShutDown()

	for i := 0; i < workers; i++ {
		go runUntil(p.worker, time.Second, stopCh)
	}

	<-stopCh
}

func (p *Processor) process() bool {
	item, shutdown := p.workQueue.Get()
	if shutdown {
		return false
	}

	// maybe need type preCheck
	job := item.(*job2.Job)
	// job execute interval
	interval := job.Interval
	// job retries count
	retries := job.RetryTimes

	_ = retry.Do(job.ExecuteFunc, retry.Attempts(retries+1))

	if !job.RunOnce {
		// put job into workQueue again with interval
		p.workQueue.AddAfter(item, interval)
	}

	job.LastFinishedTime = time.Now()
	_ = p.updateJobMetadata(job)

	return true
}

func (p *Processor) worker() {
	// if queue is empty process will return false
	for p.process() {
	}
}

func (p *Processor) saveJobMetadata(job *job2.Job) error {
	_, err := p.store.CreateJob(job)
	if err != nil {
		return err
	}
	return nil
}

func (p *Processor) updateJobMetadata(job *job2.Job) error {
	if err := p.store.UpdateJob(job.ID, job); err != nil {
		return err
	}
	return nil
}

// runUntil running f every period until stopCh
func runUntil(f func(), period time.Duration, stopCh <-chan struct{}) {
	t := time.NewTicker(period)

	for {
		select {
		case <-stopCh:
			return
		default:
		}

		f()

		select {
		case <-t.C:
		case <-stopCh:
			return
		}
	}
}
