package models

import (
	"time"
)

type User struct {
	ID        int64      `json:"id"`
	Email     string     `json:"email"`
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	DOB       *time.Time `json:"dob"`
	Phone     string     `json:"phone"`
	Password  string     `json:"password"`
}
