package validator

import (
	"errors"
	"expenses/internal/models"
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

// ValidateStatementWithOptions validates statement upload with unified ParseOptions
func (v *StatementValidator) ValidateStatementWithOptions(accountId int64, file multipart.File, header *multipart.FileHeader, options models.ParseOptions) error {
	// First validate the basic statement upload
	if err := v.ValidateStatementUpload(accountId, file, header); err != nil {
		return err
	}

	// If custom mappings are provided, validate them
	if options.HasCustomMappings() {
		customValidator := NewCustomImportValidator()
		if err := customValidator.ValidateColumnMappings(options.Mappings); err != nil {
			return err
		}
	}

	// Validate skip rows if provided
	if options.HasRowSkipping() {
		if options.SkipRows < 0 {
			return errors.New("skip_rows must be non-negative")
		}
		// Note: We can't validate against total rows here since we haven't parsed the file yet
		// This validation will be done during parsing
	}

	return nil
}
