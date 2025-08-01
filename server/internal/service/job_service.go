package service

import (
	"expenses/internal/models"
	"expenses/internal/repository"

	"github.com/gin-gonic/gin"
)

type JobServiceInterface interface {
	GetJobById(c *gin.Context, jobId int64, userId int64) (models.JobResponse, error)
	ListJobs(c *gin.Context, userId int64, query models.JobListQuery) (models.PaginatedJobResponse, error)
}

type jobService struct {
	jobRepo repository.JobRepositoryInterface
}

func NewJobService(jobRepo repository.JobRepositoryInterface) JobServiceInterface {
	return &jobService{
		jobRepo: jobRepo,
	}
}

func (s *jobService) GetJobById(c *gin.Context, jobId int64, userId int64) (models.JobResponse, error) {
	return s.jobRepo.GetJobByID(c, jobId, userId)
}

func (s *jobService) ListJobs(c *gin.Context, userId int64, query models.JobListQuery) (models.PaginatedJobResponse, error) {
	jobs, err := s.jobRepo.ListJobs(c, userId, query)
	if err != nil {
		return models.PaginatedJobResponse{}, err
	}

	total, err := s.jobRepo.CountJobs(c, userId, query)
	if err != nil {
		return models.PaginatedJobResponse{}, err
	}

	return models.PaginatedJobResponse{
		Jobs:     jobs,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}, nil
}
