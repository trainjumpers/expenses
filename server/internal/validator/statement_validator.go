package validator

import (
	"errors"
	apierrors "expenses/internal/errors"
	"strings"
)

type StatementValidator struct{}

func NewStatementValidator() *StatementValidator {
	return &StatementValidator{}
}

func (v *StatementValidator) ValidateStatementUpload(accountId int64, fileBytes []byte, fileName string) error {
	if accountId <= 0 {
		return apierrors.NewStatementBadRequestError(errors.New("invalid account id"))
	}
	if len(fileBytes) == 0 {
		return apierrors.NewStatementBadRequestError(errors.New("file is required"))
	}
	if strings.TrimSpace(fileName) == "" {
		return apierrors.NewStatementBadRequestError(errors.New("filename cannot be empty"))
	}
	if len(fileBytes) > 256*1024 {
		return apierrors.NewStatementBadRequestError(errors.New("file size must be less than 256KB"))
	}
	trimmedFileName := strings.ToLower(strings.TrimSpace(fileName))
	if !strings.HasSuffix(trimmedFileName, ".csv") && !strings.HasSuffix(trimmedFileName, ".xls") && !strings.HasSuffix(trimmedFileName, ".xlsx") && !strings.HasSuffix(trimmedFileName, ".txt") {
		return apierrors.NewStatementBadRequestError(errors.New("file must be CSV or Excel format (.csv, .xls, .xlsx, .txt)"))
	}
	return nil
}

func (v *StatementValidator) ValidateStatementPreview(fileBytes []byte, fileName string, skipRows int, rowSize int) error {
	if len(fileBytes) == 0 {
		return apierrors.NewStatementBadRequestError(errors.New("file is required"))
	}
	// Reject file bytes that are only whitespace
	if strings.TrimSpace(string(fileBytes)) == "" {
		return apierrors.NewStatementBadRequestError(errors.New("file content cannot be only whitespace"))
	}
	if strings.TrimSpace(fileName) == "" {
		return apierrors.NewStatementBadRequestError(errors.New("filename cannot be empty"))
	}
	if len(fileBytes) > 256*1024 {
		return apierrors.NewStatementBadRequestError(errors.New("file size must be less than 256KB"))
	}
	trimmedFileName := strings.ToLower(strings.TrimSpace(fileName))
	if !strings.HasSuffix(trimmedFileName, ".csv") && !strings.HasSuffix(trimmedFileName, ".xls") && !strings.HasSuffix(trimmedFileName, ".xlsx") {
		return apierrors.NewStatementBadRequestError(errors.New("file must be CSV or Excel format (.csv, .xls, .xlsx)"))
	}
	for _, ext := range []string{".csv", ".xls", ".xlsx"} {
		if strings.HasSuffix(trimmedFileName, ext) && len(strings.TrimSpace(strings.TrimSuffix(trimmedFileName, ext))) == 0 {
			return apierrors.NewStatementBadRequestError(errors.New("filename cannot be only extension"))
		}
	}
	if skipRows < 0 {
		return apierrors.NewStatementBadRequestError(errors.New("skipRows cannot be negative"))
	}
	if rowSize <= 0 {
		return apierrors.NewStatementBadRequestError(errors.New("rowSize must be positive"))
	}
	return nil
}
