package job

import "time"

// IsDeleted logic delete flag
type IsDeleted uint8

const (
	// IsDeletedUnknown
	IsDeletedUnknown IsDeleted = 0
	// IsDeletedTrue
	IsDeletedTrue IsDeleted = 1
	// IsDeletedFalse
	IsDeletedFalse IsDeleted = 2
)

const (
	defaultRetryTimes = 0
)

type Job struct {
	// Id
	ID uint `gorm:"primaryKey"`
	// job Name
	Name string
	// RetryTimes retry when run job failed
	RetryTimes uint
	// SetRunOnce indicated just run once times
	RunOnce bool
	// Interval job run Interval,it worked when SetRunOnce is true
	Interval time.Duration
	// executeFunc user defined job logic
	ExecuteFunc func() error `gorm:"-"`
	// createdTime
	CreatedTime time.Time
	// updatedTime
	UpdatedTime time.Time
	// lastFinishedTime
	LastFinishedTime time.Time
	// IsDelete
	IsDelete IsDeleted
}

// NewJob new a job
func NewJob(name string) *Job {
	return &Job{
		RunOnce:    true,
		RetryTimes: defaultRetryTimes,
		Name:       name,
	}
}

// AddFunc add user defined job logic
func (j *Job) AddFunc(f func() error) *Job {
	j.ExecuteFunc = f
	return j
}

// SetRunOnce whether to run only once
func (j *Job) SetRunOnce(once bool) *Job {
	j.RunOnce = once
	return j
}

// SetInterval set Interval
func (j *Job) SetInterval(interval time.Duration) *Job {
	j.Interval = interval
	return j
}

// SetRetry set retry
func (j *Job) SetRetry(times uint) *Job {
	j.RetryTimes = times
	return j
}
