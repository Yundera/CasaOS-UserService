package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/IceWhaleTech/CasaOS-UserService/pkg/utils/encryption"
	"github.com/IceWhaleTech/CasaOS-UserService/service/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	// TokenTypeMagicLink for passwordless login
	TokenTypeMagicLink = "magic_link"
	// TokenTypePasswordReset for password reset
	TokenTypePasswordReset = "password_reset"

	// MagicLinkExpiry is the expiration time for magic link tokens (10 minutes)
	MagicLinkExpiry = 10 * time.Minute
	// PasswordResetExpiry is the expiration time for password reset tokens (15 minutes)
	PasswordResetExpiry = 15 * time.Minute

	// RateLimitWindow is the time window for rate limiting (1 hour)
	RateLimitWindow = 1 * time.Hour
	// RateLimitMax is the maximum number of successful requests per IP per hour (typos don't count)
	RateLimitMax = 4
)

type AuthTokenService interface {
	CreateAuthToken(userID int, email, tokenType, ipAddress string) (*model.AuthTokenModel, string, string, error)
	ValidateToken(token string) (*model.AuthTokenModel, error)
	ValidateCode(email, code, tokenType string) (*model.AuthTokenModel, error)
	ConsumeToken(id int) error
	CleanupExpiredTokens() error
	CheckRateLimit(ipAddress, tokenType string) error
}

type authTokenService struct {
	db *gorm.DB
}

// CreateAuthToken creates a new auth token with a random token and code
// Returns the auth token model, plaintext token, and plaintext code
func (a *authTokenService) CreateAuthToken(userID int, email, tokenType, ipAddress string) (*model.AuthTokenModel, string, string, error) {
	// Generate cryptographically secure random token (32 bytes = 256 bits)
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		logger.Error("Failed to generate random token", zap.Error(err))
		return nil, "", "", err
	}
	plaintextToken := hex.EncodeToString(tokenBytes)

	// Generate 6-character alphanumeric code (uppercase)
	codeChars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	codeBytes := make([]byte, 6)
	if _, err := rand.Read(codeBytes); err != nil {
		logger.Error("Failed to generate random code", zap.Error(err))
		return nil, "", "", err
	}
	plaintextCode := ""
	for _, b := range codeBytes {
		plaintextCode += string(codeChars[int(b)%len(codeChars)])
	}

	// Hash token and code with MD5 for storage
	hashedToken := encryption.GetMD5ByStr(plaintextToken)
	hashedCode := encryption.GetMD5ByStr(plaintextCode)

	// Determine expiration time based on token type
	var expiresAt time.Time
	if tokenType == TokenTypeMagicLink {
		expiresAt = time.Now().Add(MagicLinkExpiry)
	} else {
		expiresAt = time.Now().Add(PasswordResetExpiry)
	}

	// Create auth token record
	authToken := model.AuthTokenModel{
		UserID:    userID,
		Token:     hashedToken,
		Code:      hashedCode,
		Type:      tokenType,
		Email:     email,
		IPAddress: ipAddress,
		Used:      false,
		ExpiresAt: expiresAt,
	}

	if err := a.db.Create(&authToken).Error; err != nil {
		logger.Error("Failed to create auth token", zap.Error(err))
		return nil, "", "", err
	}

	logger.Info("Created auth token", zap.Int("token_id", authToken.ID), zap.String("type", tokenType), zap.String("email", email))

	return &authToken, plaintextToken, plaintextCode, nil
}

// ValidateToken validates a token and returns the auth token record if valid
func (a *authTokenService) ValidateToken(token string) (*model.AuthTokenModel, error) {
	// Get all non-expired, unused tokens
	var tokens []model.AuthTokenModel
	if err := a.db.Where("used = ? AND expires_at > ?", false, time.Now()).Find(&tokens).Error; err != nil {
		logger.Error("Failed to query auth tokens", zap.Error(err))
		return nil, err
	}

	// Check each token with MD5 comparison
	hashedToken := encryption.GetMD5ByStr(token)
	for _, authToken := range tokens {
		if authToken.Token == hashedToken {
			logger.Info("Token validated successfully", zap.Int("token_id", authToken.ID))
			return &authToken, nil
		}
	}

	logger.Error("Invalid or expired token")
	return nil, errors.New("invalid or expired token")
}

// ValidateCode validates a 6-digit code for a specific email and token type
func (a *authTokenService) ValidateCode(email, code, tokenType string) (*model.AuthTokenModel, error) {
	// Normalize code to uppercase
	code = strings.ToUpper(code)

	// Get all non-expired, unused tokens for this email and type
	var tokens []model.AuthTokenModel
	if err := a.db.Where("email = ? AND type = ? AND used = ? AND expires_at > ?",
		email, tokenType, false, time.Now()).Find(&tokens).Error; err != nil {
		logger.Error("Failed to query auth tokens", zap.Error(err))
		return nil, err
	}

	// Check each code with MD5 comparison
	hashedCode := encryption.GetMD5ByStr(code)
	for _, authToken := range tokens {
		if authToken.Code == hashedCode {
			logger.Info("Code validated successfully", zap.Int("token_id", authToken.ID), zap.String("email", email))
			return &authToken, nil
		}
	}

	logger.Error("Invalid or expired code", zap.String("email", email))
	return nil, errors.New("invalid or expired code")
}

// ConsumeToken marks a token as used (one-time use enforcement)
func (a *authTokenService) ConsumeToken(id int) error {
	if err := a.db.Model(&model.AuthTokenModel{}).Where("id = ?", id).Update("used", true).Error; err != nil {
		logger.Error("Failed to consume token", zap.Error(err), zap.Int("token_id", id))
		return err
	}

	logger.Info("Token consumed", zap.Int("token_id", id))
	return nil
}

// CleanupExpiredTokens removes expired tokens from the database
func (a *authTokenService) CleanupExpiredTokens() error {
	result := a.db.Where("expires_at < ?", time.Now()).Delete(&model.AuthTokenModel{})
	if result.Error != nil {
		logger.Error("Failed to cleanup expired tokens", zap.Error(result.Error))
		return result.Error
	}

	if result.RowsAffected > 0 {
		logger.Info("Cleaned up expired tokens", zap.Int64("count", result.RowsAffected))
	}

	return nil
}

// CheckRateLimit checks if the IP has exceeded the rate limit for the given token type
func (a *authTokenService) CheckRateLimit(ipAddress, tokenType string) error {
	// Count tokens created from this IP in the rate limit window
	var count int64
	cutoff := time.Now().Add(-1 * RateLimitWindow)

	if err := a.db.Model(&model.AuthTokenModel{}).
		Where("ip_address = ? AND type = ? AND created_at > ?", ipAddress, tokenType, cutoff).
		Count(&count).Error; err != nil {
		logger.Error("Failed to check rate limit", zap.Error(err))
		return err
	}

	if count >= RateLimitMax {
		logger.Error("Rate limit exceeded", zap.String("ip", ipAddress), zap.String("type", tokenType), zap.Int64("count", count))
		return errors.New("too many requests - please wait an hour before trying again")
	}

	return nil
}

// NewAuthTokenService creates a new auth token service instance
func NewAuthTokenService(db *gorm.DB) AuthTokenService {
	return &authTokenService{
		db: db,
	}
}
