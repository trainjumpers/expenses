package models

import "time"

type Session struct {
	Id        int64
	UserId    int64
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
}
