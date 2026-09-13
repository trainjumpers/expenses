package parser

import (
	"archive/zip"
	"bytes"
	"fmt"
)

// MaxWorkbookUncompressedBytes caps the total uncompressed size of an xlsx
// archive so a zip bomb cannot expand unbounded in memory during parsing.
const MaxWorkbookUncompressedBytes = 100 * 1024 * 1024

// ValidateWorkbookSize rejects xlsx archives whose declared uncompressed size
// exceeds MaxWorkbookUncompressedBytes.
func ValidateWorkbookSize(fileBytes []byte) error {
	return validateWorkbookSize(fileBytes, MaxWorkbookUncompressedBytes)
}

func validateWorkbookSize(fileBytes []byte, maxUncompressedBytes uint64) error {
	reader, err := zip.NewReader(bytes.NewReader(fileBytes), int64(len(fileBytes)))
	if err != nil {
		return fmt.Errorf("invalid xlsx archive: %w", err)
	}

	var total uint64
	for _, file := range reader.File {
		total += file.UncompressedSize64
		if total > maxUncompressedBytes {
			return fmt.Errorf("workbook expands beyond the %d byte limit", maxUncompressedBytes)
		}
	}
	return nil
}
