package entities

type UserOutput struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Name string    `json:"name"`
}

type CreateUserInput struct {
	Email     string    `json:"email" binding:"required,email"`
	Name string    `json:"name" binding:"required"`
	Password  string    `json:"password" binding:"required,min=6"`
}

type UpdateUserInput struct {
	Name string    `json:"name,omitempty"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UpdateUserPasswordInput struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}
