package entities

import "time"

type UserOutput struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	DOB       time.Time `json:"dob"`
	Phone     string    `json:"phone"`
}

type CreateUserInput struct {
	Email     string    `json:"email" binding:"required"`
	FirstName string    `json:"first_name" binding:"required"`
	LastName  string    `json:"last_name" binding:"required"`
	DOB       time.Time `json:"dob" binding:"required"`
	Phone     string    `json:"phone" binding:"required"`
	Password  string    `json:"password" binding:"required"`
}

type UpdateUserInput struct {
	Email     string    `json:"email,omitempty"`
	FirstName string    `json:"first_name,omitempty"`
	LastName  string    `json:"last_name,omitempty"`
	DOB       time.Time `json:"dob,omitempty"`
	Phone     string    `json:"phone,omitempty"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}
