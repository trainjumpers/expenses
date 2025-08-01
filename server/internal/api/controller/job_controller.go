package controller

import (
	"expenses/internal/config"
	"expenses/internal/models"
	"expenses/internal/service"
	"expenses/pkg/logger"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type JobController struct {
	*BaseController
	jobService service.JobServiceInterface
}

func NewJobController(cfg *config.Config, jobService service.JobServiceInterface) *JobController {
	return &JobController{
		BaseController: NewBaseController(cfg),
		jobService:     jobService,
	}
}

func (j *JobController) GetJobById(ctx *gin.Context) {
	userId := j.GetAuthenticatedUserId(ctx)
	jobId, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		logger.Errorf("Failed to parse job_id: %v", err)
		j.SendError(ctx, http.StatusBadRequest, "Invalid job_id")
		return
	}

	logger.Infof("Fetching job %d for user %d", jobId, userId)
	job, err := j.jobService.GetJobById(ctx, jobId, userId)
	if err != nil {
		logger.Errorf("Error fetching job: %v", err)
		j.HandleError(ctx, err)
		return
	}

	logger.Infof("Successfully fetched job %d for user %d", jobId, userId)
	j.SendSuccess(ctx, http.StatusOK, "Job fetched successfully", job)
}

func (j *JobController) ListJobs(ctx *gin.Context) {
	userId := j.GetAuthenticatedUserId(ctx)
	logger.Infof("Fetching jobs for user %d", userId)

	var query models.JobListQuery
	if err := j.BindQuery(ctx, &query); err != nil {
		return // Error already handled by BindQuery
	}

	// Set defaults if not provided
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}

	resp, err := j.jobService.ListJobs(ctx, userId, query)
	if err != nil {
		logger.Errorf("Error fetching jobs: %v", err)
		j.HandleError(ctx, err)
		return
	}

	logger.Infof("Successfully fetched %d jobs for user %d", len(resp.Jobs), userId)
	j.SendSuccess(ctx, http.StatusOK, "Jobs fetched successfully", resp)
}
