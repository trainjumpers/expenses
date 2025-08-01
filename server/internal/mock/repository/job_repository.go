package mock_repository

import (
	"errors"
	"expenses/internal/models"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type MockJobRepository struct {
	mu     sync.Mutex
	jobs   map[int64]models.JobResponse
	nextId int64
}

func NewMockJobRepository() *MockJobRepository {
	return &MockJobRepository{
		jobs:   make(map[int64]models.JobResponse),
		nextId: 1,
	}
}

func (m *MockJobRepository) CreateJob(c *gin.Context, input models.CreateJobInput) (models.JobResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	job := models.JobResponse{
		Id:          m.nextId,
		JobType:     input.JobType,
		ReferenceId: input.ReferenceId,
		CreatedBy:   input.CreatedBy,
		Status:      input.Status,
		Message:     input.Message,
		Metadata:    input.Metadata,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	m.jobs[m.nextId] = job
	m.nextId++
	return job, nil
}

func (m *MockJobRepository) UpdateJobStatus(c *gin.Context, jobId int64, input models.UpdateJobStatusInput) (models.JobResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	job, exists := m.jobs[jobId]
	if !exists {
		return models.JobResponse{}, errors.New("job not found")
	}

	job.Status = input.Status
	if input.Message != nil {
		job.Message = input.Message
	}
	if input.StartedAt != nil {
		job.StartedAt = input.StartedAt
	}
	if input.CompletedAt != nil {
		job.CompletedAt = input.CompletedAt
	}
	job.UpdatedAt = time.Now()

	m.jobs[jobId] = job
	return job, nil
}

func (m *MockJobRepository) GetJobByID(c *gin.Context, jobId int64, userId int64) (models.JobResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	job, exists := m.jobs[jobId]
	if !exists || job.CreatedBy != userId {
		return models.JobResponse{}, errors.New("job not found")
	}

	return job, nil
}

func (m *MockJobRepository) ListJobs(c *gin.Context, userId int64, query models.JobListQuery) ([]models.JobResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var result []models.JobResponse
	for _, job := range m.jobs {
		if job.CreatedBy != userId {
			continue
		}
		if query.JobType != nil && job.JobType != *query.JobType {
			continue
		}
		if query.Status != nil && job.Status != *query.Status {
			continue
		}
		result = append(result, job)
	}

	// Simple pagination
	start := (query.Page - 1) * query.PageSize
	end := start + query.PageSize
	if start >= len(result) {
		return []models.JobResponse{}, nil
	}
	if end > len(result) {
		end = len(result)
	}

	return result[start:end], nil
}

func (m *MockJobRepository) CountJobs(c *gin.Context, userId int64, query models.JobListQuery) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	count := 0
	for _, job := range m.jobs {
		if job.CreatedBy != userId {
			continue
		}
		if query.JobType != nil && job.JobType != *query.JobType {
			continue
		}
		if query.Status != nil && job.Status != *query.Status {
			continue
		}
		count++
	}

	return count, nil
}
