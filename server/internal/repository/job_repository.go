package repository

import (
	"errors"
	"expenses/internal/config"
	"expenses/internal/database/helper"
	database "expenses/internal/database/manager"
	errorsPkg "expenses/internal/errors"
	"expenses/internal/models"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type JobRepositoryInterface interface {
	CreateJob(c *gin.Context, input models.CreateJobInput) (models.JobResponse, error)
	UpdateJobStatus(c *gin.Context, jobId int64, input models.UpdateJobStatusInput) (models.JobResponse, error)
	GetJobByID(c *gin.Context, jobId int64, userId int64) (models.JobResponse, error)
	ListJobs(c *gin.Context, userId int64, query models.JobListQuery) ([]models.JobResponse, error)
	CountJobs(c *gin.Context, userId int64, query models.JobListQuery) (int, error)
}

type JobRepository struct {
	db        database.DatabaseManager
	schema    string
	tableName string
}

func NewJobRepository(db database.DatabaseManager, cfg *config.Config) JobRepositoryInterface {
	return &JobRepository{
		db:        db,
		schema:    cfg.DBSchema,
		tableName: "job",
	}
}

func (r *JobRepository) CreateJob(c *gin.Context, input models.CreateJobInput) (models.JobResponse, error) {
	var job models.JobResponse
	query, values, ptrs, err := helper.CreateInsertQuery(&input, &job, r.tableName, r.schema)
	if err != nil {
		return job, errorsPkg.NewJobRepositoryError("failed to create job", err)
	}
	err = r.db.FetchOne(c, query, values...).Scan(ptrs...)
	if err != nil {
		return job, errorsPkg.NewJobRepositoryError("failed to create job", err)
	}
	return job, nil
}

func (r *JobRepository) UpdateJobStatus(c *gin.Context, jobId int64, input models.UpdateJobStatusInput) (models.JobResponse, error) {
	fieldsClause, argValues, argIndex, err := helper.CreateUpdateParams(&input)
	if err != nil {
		return models.JobResponse{}, errorsPkg.NewJobRepositoryError("failed to prepare update", err)
	}
	var job models.JobResponse
	ptrs, dbFields, err := helper.GetDbFieldsFromObject(&job)
	if err != nil {
		return job, errorsPkg.NewJobRepositoryError("failed to prepare update", err)
	}
	query := fmt.Sprintf(`UPDATE %s.%s SET %s WHERE id = $%d RETURNING %s;`, r.schema, r.tableName, fieldsClause, argIndex, strings.Join(dbFields, ", "))
	argValues = append(argValues, jobId)
	err = r.db.FetchOne(c, query, argValues...).Scan(ptrs...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return job, errorsPkg.NewJobNotFoundError(err)
		}
		return job, errorsPkg.NewJobRepositoryError("failed to update job", err)
	}
	return job, nil
}

func (r *JobRepository) GetJobByID(c *gin.Context, jobId int64, userId int64) (models.JobResponse, error) {
	var job models.JobResponse
	ptrs, dbFields, err := helper.GetDbFieldsFromObject(&job)
	if err != nil {
		return job, err
	}

	query := fmt.Sprintf(`
		SELECT %s
		FROM %s.%s
		WHERE id = $1 AND created_by = $2 AND deleted_at IS NULL`,
		strings.Join(dbFields, ", "), r.schema, r.tableName)
	err = r.db.FetchOne(c, query, jobId, userId).Scan(ptrs...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return job, errorsPkg.NewJobNotFoundError(err)
		}
		return job, errorsPkg.NewJobRepositoryError("failed to get job", err)
	}
	return job, nil
}

func (r *JobRepository) ListJobs(c *gin.Context, userId int64, query models.JobListQuery) ([]models.JobResponse, error) {
	jobs := make([]models.JobResponse, 0)
	var job models.JobResponse
	ptrs, dbFields, err := helper.GetDbFieldsFromObject(&job)
	if err != nil {
		return jobs, err
	}

	whereClause, args := r.buildWhereClause(userId, query)
	limit := query.PageSize
	offset := (query.Page - 1) * query.PageSize

	sqlQuery := fmt.Sprintf(`
		SELECT %s
		FROM %s.%s
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`,
		strings.Join(dbFields, ", "), r.schema, r.tableName, whereClause, len(args)+1, len(args)+2)

	args = append(args, limit, offset)
	rows, err := r.db.FetchAll(c, sqlQuery, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return jobs, nil
		}
		return jobs, errorsPkg.NewJobRepositoryError("failed to list jobs", err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(ptrs...)
		if err != nil {
			return jobs, errorsPkg.NewJobRepositoryError("failed to scan job row", err)
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func (r *JobRepository) CountJobs(c *gin.Context, userId int64, query models.JobListQuery) (int, error) {
	whereClause, args := r.buildWhereClause(userId, query)
	sqlQuery := fmt.Sprintf(`SELECT COUNT(*) FROM %s.%s WHERE %s`, r.schema, r.tableName, whereClause)
	var count int
	err := r.db.FetchOne(c, sqlQuery, args...).Scan(&count)
	if err != nil {
		return 0, errorsPkg.NewJobRepositoryError("failed to count jobs", err)
	}
	return count, nil
}

func (r *JobRepository) buildWhereClause(userId int64, query models.JobListQuery) (string, []interface{}) {
	conditions := []string{"created_by = $1", "deleted_at IS NULL"}
	args := []interface{}{userId}
	argIndex := 2

	if query.JobType != nil {
		conditions = append(conditions, fmt.Sprintf("job_type = $%d", argIndex))
		args = append(args, *query.JobType)
		argIndex++
	}

	if query.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *query.Status)
		argIndex++
	}

	return strings.Join(conditions, " AND "), args
}
