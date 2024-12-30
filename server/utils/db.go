package utils

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func GetPGSchema() string {
	schema := os.Getenv("PGSCHEMA")
	if schema == "" {
		panic("PGSCHEMA environment variable is not set")
	}
	return schema
}

func CheckForeignKey(err error, table string, fkKey string) bool {
	return strings.Contains(err.Error(), table) && strings.Contains(err.Error(), fkKey)
}

func CreateUpdateParamsQuery(fields map[string]interface{}) (string, []interface{}, int, error) {
	fieldsClause := ""
	argIndex := 1
	argValues := make([]interface{}, 0)
	for k, v := range fields {
		if v == "" || v == int64(0) {
			continue
		}

		fieldsClause += k + " = $" + strconv.FormatInt(int64(argIndex), 10) + ", "
		argIndex++
		argValues = append(argValues, v)
	}
	fieldsClause = strings.TrimSuffix(fieldsClause, ", ")
	if fieldsClause == "" {
		return fieldsClause, argValues, argIndex, fmt.Errorf("no fields to update")
	}
	return fieldsClause, argValues, argIndex, nil
}