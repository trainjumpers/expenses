package validator

import (
	"errors"
	"mime/multipart"
	"strconv"
	"strings"
)

type StatementValidator struct{}

func NewStatementValidator() *StatementValidator {
	return &StatementValidator{}
}

func (v *StatementValidator) ValidateStatementUpload(accountId int64, file multipart.File, header *multipart.FileHeader) error {
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
	filename := strings.ToLower(header.Filename)
	if !strings.HasSuffix(filename, ".csv") && !strings.HasSuffix(filename, ".xls") && !strings.HasSuffix(filename, ".xlsx") {
		return errors.New("file must be CSV or Excel format (.csv, .xls, .xlsx)")
	}

	return nil
}

func (v *StatementValidator) ValidatePaginationParams(limitStr, offsetStr string) (int, int, error) {
	limit := 10 // default
	offset := 0 // default

	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil {
			return 0, 0, errors.New("invalid limit parameter")
		}
		if parsedLimit <= 0 || parsedLimit > 100 {
			return 0, 0, errors.New("limit must be between 1 and 100")
		}
		limit = parsedLimit
	}

	if offsetStr != "" {
		parsedOffset, err := strconv.Atoi(offsetStr)
		if err != nil {
			return 0, 0, errors.New("invalid offset parameter")
		}
		if parsedOffset < 0 {
			return 0, 0, errors.New("offset must be non-negative")
		}
		offset = parsedOffset
	}

	return limit, offset, nil
}
