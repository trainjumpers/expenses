package utils

import (
	"errors"
	"reflect"
)

// extractFields extracts pointers, values, and column names for exported struct fields.
func ExtractFields(obj interface{}, skipNull bool) ([]interface{}, []interface{}, []string, error) {
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
	if !v.IsValid() {
		return true
	}
	if v.Kind() == reflect.Ptr {
		return v.IsNil()
	}
	return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}

func ConvertStruct(src interface{}, dst interface{}) {
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
