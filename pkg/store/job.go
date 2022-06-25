package store

import (
	"github.com/jinzhu/gorm"

	"github.com/mosesyou/scheduler-demo/pkg/job"
)

const (
	// jobTableName job table name
	jobTableName = "scheduler_job"
)

var _ JobStore = (*jobStore)(nil)

// JobStore
type JobStore interface {

	// CreateJob
	CreateJob(job *job.Job) (*job.Job, error)

	// UpdateJob
	UpdateJob(id uint, job *job.Job) error

	// RemoveJob
	RemoveJob(id uint) error
}

// jobStore
type jobStore struct {
	db *gorm.DB
}

func NewJobStore(db *gorm.DB) JobStore {
	return &jobStore{
		db: db,
	}
}

// CreateJob create a job
func (s *jobStore) CreateJob(job *job.Job) (*job.Job, error) {
	if err := s.db.Table(jobTableName).Create(job).Error; err != nil {
		return nil, err
	}
	return job, nil
}

// UpdateJob update a job
func (s *jobStore) UpdateJob(id uint, job *job.Job) error {
	return s.db.Table(jobTableName).
		Where("id = ?", id).Updates(job).Error
}

// RemoveJob remove a job,logic delete
func (s *jobStore) RemoveJob(id uint) error {
	var item = map[string]interface{}{
		"is_delete": job.IsDeletedTrue,
	}

	return s.db.Table(jobTableName).Where("id = ?", id).Updates(item).Error
}
