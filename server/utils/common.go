package utils

import (
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func GetPGSchema() string {
	schema := os.Getenv("PGSCHEMA")
	if schema == "" {
		panic("PGSCHEMA environment variable is not set")
	}
	return schema
}

// hashPassword hashes the password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// checkPasswordHash checks if the password matches the hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func CheckForeignKey(err error, table string, fkKey string) bool {
	return strings.Contains(err.Error(), "fk_"+table) && strings.Contains(err.Error(), fkKey)
}
