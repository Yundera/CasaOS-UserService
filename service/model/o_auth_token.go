package model

import "time"

// AuthTokenModel represents authentication tokens for magic link and password reset
type AuthTokenModel struct {
	ID        int       `gorm:"column:id;primary_key" json:"id"`
	UserID    int       `gorm:"column:user_id;index" json:"user_id"`     // Foreign key to o_users
	Token     string    `gorm:"column:token;uniqueIndex" json:"token"`   // MD5-hashed token
	Code      string    `gorm:"column:code" json:"code"`                 // 6-digit code (MD5-hashed)
	Type      string    `gorm:"column:type" json:"type"`                 // "magic_link" or "password_reset"
	Email     string    `gorm:"column:email;index" json:"email"`         // Email associated with request
	IPAddress string    `gorm:"column:ip_address;index" json:"ip_address"` // IP address for rate limiting
	Used      bool      `gorm:"column:used;default:false" json:"used"`
	ExpiresAt time.Time `gorm:"column:expires_at;index" json:"expires_at"`
	CreatedAt time.Time `gorm:"<-:create;autoCreateTime" json:"created_at"`
}

func (a *AuthTokenModel) TableName() string {
	return "o_auth_tokens"
}
