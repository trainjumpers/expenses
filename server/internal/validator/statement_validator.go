package validator

import (
	"errors"
	"mime/multipart"
	"strings"
)

type StatementValidator struct{}

func NewStatementValidator() *StatementValidator {
	return &StatementValidator{}
}

func (v *StatementValidator) ValidateStatementUpload(accountId int64, file multipart.File, header *multipart.FileHeader) error {
	// Validate accountId
	if accountId <= 0 {
		return errors.New("invalid account id")
	}

	// Validate file
	if file == nil || header == nil {
		return errors.New("file is required")
	}

	// Validate filename is not empty
	if strings.TrimSpace(header.Filename) == "" {
		return errors.New("filename cannot be empty")
	}

	// Validate file size (256KB max)
	if header.Size > 256*1024 {
		return errors.New("file size must be less than 256KB")
	}

	// Validate file type
	filename := strings.ToLower(strings.TrimSpace(header.Filename))
	if !strings.HasSuffix(filename, ".csv") && !strings.HasSuffix(filename, ".xls") && !strings.HasSuffix(filename, ".xlsx") {
		return errors.New("file must be CSV or Excel format (.csv, .xls, .xlsx)")
	}

	return nil
}
