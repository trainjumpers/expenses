package models

import "time"

type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         UserResponse `json:"user"`
}

type RefreshTokenData struct {
	UserId int64
	Email  string
	Expiry time.Time
}
