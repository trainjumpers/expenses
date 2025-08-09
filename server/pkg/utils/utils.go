package utils

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// extractFields extracts pointers, values, and column names for exported struct fields.
func ExtractFields(obj any, skipNull bool) ([]any, []any, []string, error) {
	if obj == nil {
		return nil, nil, nil, errors.New("extractFields: obj is nil")
	}
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return nil, nil, nil, errors.New("extractFields: obj must be a non-nil pointer to a struct")
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return nil, nil, nil, errors.New("extractFields: obj must be a pointer to a struct")
	}
	t := v.Type()
	var ptrs []any
	var values []any
	var fields []string
	for i := range t.NumField() {
		field := v.Field(i)
		fieldType := t.Field(i)
		// Skip unexported fields
		if fieldType.PkgPath != "" {
			continue
		}
		if skipNull && IsZeroValue(field) {
			continue
		}
		ptrs = append(ptrs, field.Addr().Interface())
		values = append(values, field.Interface())
		fields = append(fields, fieldType.Name)
	}
	return ptrs, values, fields, nil
}

func IsZeroValue(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}
	if v.Kind() == reflect.Ptr {
		return v.IsNil()
	}
	return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}

func ConvertStruct(src any, dst any) {
	if src == nil || dst == nil {
		return
	}
	srcReflect := reflect.ValueOf(src)
	if srcReflect.Kind() != reflect.Ptr || srcReflect.IsNil() {
		return
	}
	dstReflect := reflect.ValueOf(dst)
	if dstReflect.Kind() != reflect.Ptr || dstReflect.IsNil() {
		return
	}
	srcVal := srcReflect.Elem()
	dstVal := dstReflect.Elem()

	if srcVal.Kind() != reflect.Struct || dstVal.Kind() != reflect.Struct {
		return
	}

	srcType := srcVal.Type()
	dstType := dstVal.Type()
	for i := 0; i < dstType.NumField(); i++ {
		dstField := dstType.Field(i)
		srcField, ok := srcType.FieldByName(dstField.Name)
		if ok && srcField.Type == dstField.Type {
			dstVal.Field(i).Set(srcVal.FieldByName(dstField.Name))
		}
	}
}

// ParseDate tries to parse a date string with multiple common layouts.
func ParseDate(dateStr string) (time.Time, error) {
	layouts := []string{
		// Standard formats
		"2006-01-02", "2006/01/02",
		// SBI formats
		"2 Jan 2006", "02 Jan 2006", "2 January 2006", "02 January 2006",
		// HDFC and ICICI formats (DD/MM/YYYY and DD/MM/YY) - prioritize these
		"02/01/2006", "2/1/2006", "02/01/06", "2/1/06",
		// US formats (MM/DD/YYYY) - these come after DD/MM to avoid ambiguity
		"01-02-2006", "01/02/2006",
		// Other common formats
		"Jan 2, 2006", "January 2, 2006",
		time.RFC3339,
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, dateStr); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// ParseFloat cleans and parses a string into a float64, removing commas and handling empty strings.
func ParseFloat(amountStr string) (float64, error) {
	// Remove commas and extra spaces
	cleanAmount := strings.ReplaceAll(amountStr, ",", "")
	cleanAmount = strings.TrimSpace(cleanAmount)

	if cleanAmount == "" {
		return 0, errors.New("empty amount string")
	}

	amount, err := strconv.ParseFloat(cleanAmount, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid amount format: %s", amountStr)
	}

	return amount, nil
}
