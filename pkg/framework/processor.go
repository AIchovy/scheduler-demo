package framework

import (
	"time"

	"github.com/avast/retry-go"

	"github.com/mosesyou/scheduler-demo/pkg/job"
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
	j := item.(*job.Job)
	// job execute interval
	interval := j.Interval
	// job retries count
	retries := j.RetryTimes

	_ = retry.Do(j.ExecuteFunc, retry.Attempts(retries+1))

	if !j.RunOnce {
		// put job into workQueue again with interval
		p.workQueue.AddAfter(item, interval)
	}

	j.LastFinishedTime = time.Now()
	_ = p.updateJobMetadata(j)

	return true
}

func (p *Processor) worker() {
	// if queue is empty process will return false
	for p.process() {
	}
}

func (p *Processor) saveJobMetadata(j *job.Job) error {
	_, err := p.store.CreateJob(j)
	if err != nil {
		return err
	}
	return nil
}

func (p *Processor) updateJobMetadata(j *job.Job) error {
	if err := p.store.UpdateJob(j.ID, j); err != nil {
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
