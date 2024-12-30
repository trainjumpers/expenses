package services

import (
	"encoding/json"
	"expenses/logger"
	"expenses/models"
	"expenses/utils"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type JobService struct {
	db     *pgxpool.Pool
	schema string
}

func NewJobService(db *pgxpool.Pool) *JobService {
	return &JobService{
		db:     db,
		schema: utils.GetPGSchema(), //unable to load as this is not inited anywhere in main, thus doesnt have access to env
	}
}

func (e *JobService) GetJobStatus(c *gin.Context, userID int64, jobID int64) (models.JobStatus, error) {
	query := fmt.Sprintf(`
	SELECT
		j.id,
		j.name,
		j.metadata,
		j.status
	FROM %[1]s.jobs
	LEFT JOIN %[1]s.job_status_mappings jsm ON j.id = jsm.job_id
	WHERE j.id = $1 AND jsm.user_id = $2
	`, e.schema)

	logger.Info("Executing query: %s", query)
	record := e.db.QueryRow(c, query, jobID, userID)
	var jobStatus models.JobStatus

	logger.Info("Successfully queried job status for user with ID: %d", userID)
	logger.Info("Scanning job status for user with ID: %d", userID)
	err := record.Scan(&jobStatus.ID, &jobStatus.Name, &jobStatus.Metadata, &jobStatus.Status)
	if err != nil {
		logger.Error("Error scanning job status for user with ID: %d", userID)
		return jobStatus, err
	}
	logger.Info("Successfully scanned job status for user with ID: %d", userID)
	return jobStatus, nil
}

func (e *JobService) CreateJob(c *gin.Context, userID int64, name string, metadata json.RawMessage) (models.JobStatus, error) {
	query := fmt.Sprintf(`
	WITH inserted_job AS (
		INSERT INTO %[1]s.jobs (name, metadata)
		VALUES ($1, $2)
		RETURNING id, name, metadata, status
	), new_mapping AS (
		INSERT INTO %[1]s.jobs_user_mapping (job_id, user_id)
		SELECT id, $3 FROM inserted_job jb
		RETURNING *
	)
	SELECT
		j.id,
		j.name,
		j.metadata,
		j.status
	FROM inserted_job j
	`, e.schema)

	logger.Info("Executing query: %s", query)
	record := e.db.QueryRow(c, query, name, metadata, userID)
	var jobStatus models.JobStatus

	logger.Info("Successfully created job status for user with ID: %d", userID)
	logger.Info("Scanning job status for user with ID: %d", userID)
	err := record.Scan(&jobStatus.ID, &jobStatus.Name, &jobStatus.Metadata, &jobStatus.Status)
	if err != nil {
		logger.Error("Error scanning job status for user with ID: %d", userID)
		return jobStatus, err
	}
	logger.Info("Successfully scanned job status for user with ID: %d", userID)
	return jobStatus, nil
}

func (e *JobService) UpdateJobStatus(c *gin.Context, userID int64, jobID int64, status string) (models.JobStatus, error) {
	query := fmt.Sprintf(`
	WITH update_job AS (
		UPDATE %[1]s.jobs j
			SET status = $1
		FROM %[1]s.jobs_user_mapping m
		WHERE j.id = $2 AND m.job_id = j.id AND m.user_id = $3
		RETURNING j.id, j.status
	)
	SELECT
		id,
		status
	FROM update_job
	`, e.schema)

	logger.Info("Executing query: %s", query)
	record := e.db.QueryRow(c, query, status, jobID, userID)
	var jobStatus models.JobStatus

	logger.Info("Updating job status for user with ID: %d and job ID: %d", userID, jobID)
	logger.Info("Scanning updated job status for user with ID: %d", userID)
	err := record.Scan(&jobStatus.ID, &jobStatus.Status)
	if err != nil {
		logger.Error("Error scanning updated job status for user with ID: %d", userID)
		return jobStatus, err
	}
	logger.Info("Successfully updated and scanned job status for user with ID: %d", userID)
	return jobStatus, nil
}
