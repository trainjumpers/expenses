package validator

import (
	"expenses/internal/errors"
	"expenses/internal/models"
	"mime/multipart"
	"path/filepath"
	"strings"
)

type CustomImportValidator struct{}

func NewCustomImportValidator() *CustomImportValidator {
	return &CustomImportValidator{}
}

// ValidateCSVFile validates the uploaded CSV/XLS file
func (v *CustomImportValidator) ValidateCSVFile(file multipart.File, header *multipart.FileHeader) error {
	// Check file size (256KB limit)
	if header.Size > 256*1024 {
		return errors.NewCSVFileTooLargeError()
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	validExtensions := []string{".csv", ".xls", ".xlsx"}
	
	isValidExt := false
	for _, validExt := range validExtensions {
		if ext == validExt {
			isValidExt = true
			break
		}
	}
	
	if !isValidExt {
		return errors.NewInvalidCSVFormatError(nil)
	}

	return nil
}

// ValidateColumnMappings validates that required fields are mapped and no duplicates exist
func (v *CustomImportValidator) ValidateColumnMappings(mappings []models.ColumnMapping) error {
	if len(mappings) == 0 {
		return errors.NewMissingRequiredFieldError("at least one mapping")
	}

	// Track mapped fields to check for duplicates
	mappedFields := make(map[string]bool)
	requiredFields := map[string]bool{
		"name":   false,
		"amount": false,
		"date":   false,
	}

	// Check for amount OR (credit AND debit)
	hasAmount := false
	hasCredit := false
	hasDebit := false

	for _, mapping := range mappings {
		// Check for duplicate mappings
		if mappedFields[mapping.TargetField] {
			return errors.NewDuplicateMappingError(mapping.TargetField)
		}
		mappedFields[mapping.TargetField] = true

		// Track required fields
		switch mapping.TargetField {
		case "name":
			requiredFields["name"] = true
		case "amount":
			hasAmount = true
			requiredFields["amount"] = true
		case "credit":
			hasCredit = true
		case "debit":
			hasDebit = true
		case "date":
			requiredFields["date"] = true
		}
	}

	// Check if we have amount OR both credit and debit
	if !hasAmount && !(hasCredit && hasDebit) {
		return errors.NewMissingRequiredFieldError("amount (or both credit and debit)")
	}

	// If we have credit/debit, mark amount requirement as satisfied
	if hasCredit && hasDebit {
		requiredFields["amount"] = true
	}

	// Check for missing required fields
	for field, mapped := range requiredFields {
		if !mapped {
			return errors.NewMissingRequiredFieldError(field)
		}
	}

	return nil
}

// ValidateSkipRows validates that skip rows count is reasonable
func (v *CustomImportValidator) ValidateSkipRows(skipRows int, totalRows int) error {
	if skipRows < 0 {
		return errors.NewInvalidCSVFormatError(nil)
	}

	if skipRows >= totalRows {
		return errors.NewInsufficientColumnsError()
	}

	return nil
}