package utils

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

var ErrNoFieldsToUpdate = errors.New("no fields to update")

func GetPGSchema() string {
	schema := os.Getenv("PG_SCHEMA")
	if schema == "" {
		panic("PG_SCHEMA environment variable is not set")
	}
	return schema
}

func CheckForeignKey(err error, table string, fkKey string) bool {
	return strings.Contains(err.Error(), table) && strings.Contains(err.Error(), fkKey)
}

// CreateUpdateParams generates an SQL update clause and values from a struct pointer.
func CreateUpdateParams(obj interface{}) (string, []interface{}, int, error) {
	_, values, dbFields, err := extractDbFields(obj, true)
	if err != nil {
		return "", nil, 0, err
	}
	if len(dbFields) == 0 {
		return "", nil, 1, ErrNoFieldsToUpdate
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
	if obj == nil {
		return nil, nil, nil, errors.New("extractDbFields: obj is nil")
	}
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return nil, nil, nil, errors.New("extractDbFields: obj must be a non-nil pointer to a struct")
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return nil, nil, nil, errors.New("extractDbFields: obj must be a pointer to a struct")
	}
	t := v.Type()
	var ptrs []interface{}
	var values []interface{}
	var dbFields []string
	for i := range t.NumField() {
		field := v.Field(i)
		fieldType := t.Field(i)
		// Skip unexported fields
		if fieldType.PkgPath != "" {
			continue
		}
		if skipNull && isZeroValue(field) {
			continue
		}
		ptrs = append(ptrs, field.Addr().Interface())
		values = append(values, field.Interface())
		dbFields = append(dbFields, toSnakeCase(fieldType.Name))
	}
	return ptrs, values, dbFields, nil
}

func isZeroValue(v reflect.Value) bool {
	if v.Kind() == reflect.Ptr {
		return v.IsNil()
	}
	return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}
