package services

import (
	"expenses/utils"

	"github.com/jackc/pgx/v5/pgxpool"
)

type StatementService struct {
	db     *pgxpool.Pool
	schema string
}

func NewStatementService(db *pgxpool.Pool) *StatementService {
	return &StatementService{
		db:     db,
		schema: utils.GetPGSchema(), //unable to load as this is not inited anywhere in main, thus doesnt have access to env
	}
}
