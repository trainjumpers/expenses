package helper

import (
	"expenses/internal/errors"
	"expenses/pkg/utils"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// GetPGSchema retrieves the PostgreSQL schema from the environment variable DB_SCHEMA.
func GetPGSchema() string {
	schema := os.Getenv("DB_SCHEMA")
	if schema == "" {
		panic("DB_SCHEMA environment variable is not set")
	}
	return schema
}

// CreateUpdateParams generates an SQL update clause and values from a struct pointer.
func CreateUpdateParams(obj interface{}) (string, []interface{}, int, error) {
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
func CreateInsertQuery(insertObj interface{}, outputObj interface{}, tableName string) (string, []interface{}, []interface{}, error) {
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
		INSERT INTO %s.%s (%s)
		VALUES (%s) RETURNING %s;`,
		GetPGSchema(), tableName, columnsClause, placeholdersClause, strings.Join(dbFields, ", "))
	return query, values, ptrs, nil
}

// GetDbFieldsFromObject returns pointers to all struct fields in order.
func GetDbFieldsFromObject(obj interface{}) ([]interface{}, []string, error) {
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

// toSnakeCase converts a PascalCase or camelCase string to snake_case.
func toSnakeCase(str string) string {
	var result []rune
	for i, r := range str {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, r)
	}
	return strings.ToLower(string(result))
}

// extractDbFields extracts pointers, values, and column names for exported struct fields.
func extractDbFields(obj interface{}, skipNull bool) ([]interface{}, []interface{}, []string, error) {
	ptrs, values, fields, err := utils.ExtractFields(obj, skipNull)
	if err != nil {
		return nil, nil, nil, err
	}
	for i, field := range fields {
		fields[i] = toSnakeCase(field)
	}
	if len(fields) == 0 {
		return nil, nil, nil, errors.NoFieldsToUpdateError()
	}
	return ptrs, values, fields, nil
}
