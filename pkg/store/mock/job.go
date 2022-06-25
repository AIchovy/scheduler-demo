package mock

import (
	"github.com/mosesyou/scheduler-demo/pkg/job"
	"github.com/mosesyou/scheduler-demo/pkg/store"
)

// JobStoreMock just use for test
type JobStoreMock struct {
}

func NewJobStoreMock() store.JobStore {
	return &JobStoreMock{}
}

// CreateJob create a job
func (s *JobStoreMock) CreateJob(job *job.Job) (*job.Job, error) {
	return job, nil
}

// UpdateJob update a job
func (s *JobStoreMock) UpdateJob(id uint, job *job.Job) error {
	return nil
}

// RemoveJob remove a job,logic delete
func (s *JobStoreMock) RemoveJob(id uint) error {
	return nil
}
