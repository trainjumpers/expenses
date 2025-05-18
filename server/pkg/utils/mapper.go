package utils

import (
	"errors"
	"reflect"
)

// extractFields extracts pointers, values, and column names for exported struct fields.
func ExtractFields(obj interface{}, skipNull bool) ([]interface{}, []interface{}, []string, error) {
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
	if v.Kind() == reflect.Ptr {
		return v.IsNil()
	}
	return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}
