package helper

import (
	"expenses/internal/errors"
	"expenses/pkg/utils"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// CreateUpdateParams generates an SQL update clause and values from a struct pointer.
func CreateUpdateParams(obj any) (string, []any, int, error) {
	_, values, dbFields, err := extractDbFields(obj, true)
	if err != nil {
		return "", nil, 0, err
	}
	if len(dbFields) == 0 {
		return "", nil, 1, errors.NoFieldsToUpdateError()
	}
	setClauses := make([]string, len(dbFields))
	for i, col := range dbFields {
		setClauses[i] = col + " = $" + strconv.Itoa(i+1)
	}
	fieldsClause := strings.Join(setClauses, ", ")
	return fieldsClause, values, len(values) + 1, nil
}

// CreateInsertQuery generates an SQL insert clause and values from a struct pointer.
func CreateInsertQuery(insertObj any, outputObj any, tableName string, schema string) (string, []any, []any, error) {
	_, values, columns, err := extractDbFields(insertObj, true)
	if err != nil {
		return "", nil, nil, err
	}
	placeholders := make([]string, len(columns))
	for i := range columns {
		placeholders[i] = "$" + strconv.Itoa(i+1)
	}
	if len(columns) == 0 {
		return "", nil, nil, fmt.Errorf("no fields to insert")
	}
	columnsClause := strings.Join(columns, ", ")
	placeholdersClause := strings.Join(placeholders, ", ")
	if len(values) == 0 {
		return "", nil, nil, fmt.Errorf("no values to insert")
	}
	ptrs, _, dbFields, err := extractDbFields(outputObj, false)
	if err != nil {
		return "", nil, nil, err
	}
	query := fmt.Sprintf(`
		INSERT INTO %s.%s (%s) VALUES (%s) RETURNING %s;`,
		schema, tableName, columnsClause, placeholdersClause, strings.Join(dbFields, ", "))
	return query, values, ptrs, nil
}

// GetDbFieldsFromObject returns pointers to all struct fields in order.
func GetDbFieldsFromObject(obj any) ([]any, []string, error) {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return nil, nil, errors.New("obj must be a non-nil pointer to a struct")
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return nil, nil, errors.New("obj must be a pointer to a struct")
	}
	ptrs, _, dbFields, err := extractDbFields(obj, false)
	if err != nil {
		return nil, nil, err
	}
	return ptrs, dbFields, nil
}

// ToSnakeCase converts a PascalCase or camelCase string to snake_case.
func ToSnakeCase(str string) string {
	// Step 1: Remove whitespaces
	str = strings.TrimSpace(str)
	if str == "" {
		return ""
	}
	// Step 2: Convert CamelCase to snake_case
	firstCap := regexp.MustCompile(`(.)([A-Z][a-z]+)`)       // First letter of the word is lowercase
	allCap := regexp.MustCompile(`([a-z0-9])([A-Z])`)        // All letters are uppercase
	letterNumber := regexp.MustCompile(`([A-Za-z])([0-9]+)`) // Letter followed by number
	numberLetter := regexp.MustCompile(`([0-9]+)([A-Za-z])`) // Number followed by letter
	snake := firstCap.ReplaceAllString(str, `${1}_${2}`)
	snake = allCap.ReplaceAllString(snake, `${1}_${2}`)
	snake = letterNumber.ReplaceAllString(snake, `${1}_${2}`)
	snake = numberLetter.ReplaceAllString(snake, `${1}_${2}`)
	// Step 3: normalize separators to underscores.
	replacer := strings.NewReplacer(
		"-", "_",
		" ", "_",
	)
	snake = replacer.Replace(snake)
	// Step 4: collapse any doubled underscores left over.
	snake = strings.ReplaceAll(snake, "__", "_")
	// Step 5: final canonical form — lower-case.
	return strings.ToLower(snake)
}

// extractDbFields extracts pointers, values, and column names for exported struct fields.
func extractDbFields(obj any, skipNull bool) ([]any, []any, []string, error) {
	ptrs, values, fields, err := utils.ExtractFields(obj, skipNull)
	if err != nil {
		return nil, nil, nil, err
	}
	for i, field := range fields {
		fields[i] = ToSnakeCase(field)
	}
	if len(fields) == 0 {
		return nil, nil, nil, errors.NoFieldsToUpdateError()
	}
	return ptrs, values, fields, nil
}
