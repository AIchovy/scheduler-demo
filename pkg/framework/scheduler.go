package framework

import (
	"context"

	"github.com/mosesyou/scheduler-demo/pkg/job"
	"github.com/mosesyou/scheduler-demo/pkg/store/mock"
)

const (
	defaultBuffer      = 10
	defaultWorkerCount = 1
)

// Config Scheduler Config
type Config struct {
	queueBufferSize uint
	workerCount     int
}

// Option User Option
type Option func(*Config)

// WorkerCount
func WorkerCount(count int) Option {
	return func(c *Config) {
		c.workerCount = count
	}
}

// QueueBuffer
func QueueBuffer(size uint) Option {
	return func(c *Config) {
		c.queueBufferSize = size
	}
}

///////////////    Scheduler   /////////////////

// Scheduler
type Scheduler struct {
	*Config
	// ctx framework context
	ctx context.Context
	// cancel context cancel func
	cancel func()
	// processor job processor
	processor *Processor
}

// NewScheduler new framework with user option
func NewScheduler(opts ...Option) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())

	// set default config
	config := &Config{
		queueBufferSize: defaultBuffer,
		workerCount:     defaultWorkerCount,
	}

	// reset user opts
	for _, opt := range opts {
		opt(config)
	}

	// new job store
	storeMock := mock.NewJobStoreMock()

	return &Scheduler{
		processor: NewProcessor(storeMock, newWorkQueue(config.queueBufferSize)),
		ctx:       ctx,
		cancel:    cancel,
		Config:    config,
	}
}

func (s *Scheduler) Start() {
	// run processor
	s.processor.Run(s.workerCount, s.ctx.Done())
}

// ScheduleJob schedule a job,it will add to queue directly
func (s *Scheduler) ScheduleJob(job *job.Job) error {
	// store job metadata before scheduling
	if err := s.processor.saveJobMetadata(job); err != nil {
		return err
	}
	s.processor.workQueue.Add(job)

	return nil
}

func (s *Scheduler) Close() {
	if s.cancel != nil {
		s.cancel()
	}
}
