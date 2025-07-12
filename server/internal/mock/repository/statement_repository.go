package mock_repository

import (
	"errors"
	"expenses/internal/models"
	"sync"

	"github.com/gin-gonic/gin"
)

type MockStatementRepository struct {
	statements           map[int64]models.StatementResponse
	nextId               int64
	mu                   sync.RWMutex
	statementTxnMappings []statementTxnMapping
}

func NewMockStatementRepository() *MockStatementRepository {
	return &MockStatementRepository{
		statements:           make(map[int64]models.StatementResponse),
		nextId:               1,
		statementTxnMappings: []statementTxnMapping{},
	}
}

func (m *MockStatementRepository) CreateStatement(c *gin.Context, input models.CreateStatementInput) (models.StatementResponse, error) {
	if input.AccountId <= 0 {
		return models.StatementResponse{}, errors.New("invalid account id")
	}
	if input.OriginalFilename == "" {
		return models.StatementResponse{}, errors.New("filename cannot be empty")
	}
	fileType := input.FileType
	if fileType != "csv" && fileType != "xls" && fileType != "xlsx" {
		return models.StatementResponse{}, errors.New("invalid file type")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	id := m.nextId
	m.nextId++
	statement := models.StatementResponse{
		Id:               id,
		AccountId:        input.AccountId,
		CreatedBy:        input.CreatedBy,
		OriginalFilename: input.OriginalFilename,
		FileType:         input.FileType,
		Status:           input.Status,
		Message:          input.Message,
	}
	m.statements[id] = statement
	return statement, nil
}

// CreateStatementTxn adds a mapping between statement and transaction for testing
func (m *MockStatementRepository) CreateStatementTxn(c *gin.Context, statementId int64, transactionId int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.statementTxnMappings = append(m.statementTxnMappings, statementTxnMapping{
		StatementId:   statementId,
		TransactionId: transactionId,
	})
	return nil
}

func (m *MockStatementRepository) UpdateStatementStatus(c *gin.Context, statementId int64, input models.UpdateStatementStatusInput) (models.StatementResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	statement, ok := m.statements[statementId]
	if !ok {
		return models.StatementResponse{}, errors.New("statement not found")
	}
	statement.Status = input.Status
	statement.Message = input.Message
	m.statements[statementId] = statement
	return statement, nil
}

func (m *MockStatementRepository) GetStatementByID(c *gin.Context, statementId int64, userId int64) (models.StatementResponse, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	statement, ok := m.statements[statementId]
	if !ok {
		return models.StatementResponse{}, errors.New("statement not found")
	}
	if statement.CreatedBy != userId {
		return models.StatementResponse{}, errors.New("unauthorized access")
	}
	return statement, nil
}

func (m *MockStatementRepository) ListStatementByUserId(c *gin.Context, userId int64, limit, offset int) ([]models.StatementResponse, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []models.StatementResponse
	for _, s := range m.statements {
		if s.CreatedBy == userId {
			result = append(result, s)
		}
	}
	// Sort by ID DESC for simplicity
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].Id < result[j].Id {
				result[i], result[j] = result[j], result[i]
			}
		}
	}
	start := offset
	end := offset + limit
	if start > len(result) {
		return []models.StatementResponse{}, nil
	}
	if end > len(result) {
		end = len(result)
	}
	return result[start:end], nil
}

func (m *MockStatementRepository) CountStatementsByUserId(c *gin.Context, userId int64) (int, error) {
	count := 0
	for _, s := range m.statements {
		if s.CreatedBy == userId {
			count++
		}
	}
	return count, nil
}
