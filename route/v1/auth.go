package v1

import (
	"fmt"
	"net/smtp"
	"os"
	"strings"

	"github.com/IceWhaleTech/CasaOS-Common/utils/common_err"
	"github.com/IceWhaleTech/CasaOS-Common/utils/jwt"
	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/IceWhaleTech/CasaOS-UserService/model"
	"github.com/IceWhaleTech/CasaOS-UserService/model/system_model"
	"github.com/IceWhaleTech/CasaOS-UserService/pkg/utils/email"
	"github.com/IceWhaleTech/CasaOS-UserService/service"
	model2 "github.com/IceWhaleTech/CasaOS-UserService/service/model"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"time"
)

// PostMagicLinkRequest requests a magic link for passwordless login
// @Summary Request magic link
// @Description Send a magic link email for passwordless login
// @Tags auth
// @Accept json
// @Produce json
// @Param email body string true "Email address"
// @Success 200 {object} model.Result
// @Router /v1/auth/magic-link/request [post]
func PostMagicLinkRequest(ctx echo.Context) error {
	json := make(map[string]string)
	ctx.Bind(&json)

	userEmail := strings.TrimSpace(json["email"])

	// Validate email format
	if len(userEmail) == 0 || !strings.Contains(userEmail, "@") {
		return ctx.JSON(common_err.CLIENT_ERROR,
			model.Result{Success: common_err.INVALID_PARAMS, Message: "Invalid email address"})
	}

	// Find user by email (matched against the USER_EMAIL env var)
	user := getUserByEmail(userEmail)
	if user.Id == 0 {
		// Email not found - wait 2 seconds to prevent timing attacks
		time.Sleep(2 * time.Second)
		return ctx.JSON(common_err.CLIENT_ERROR,
			model.Result{Success: common_err.INVALID_PARAMS, Message: "Email address not found. Please check and try again."})
	}

	// Get client IP address
	clientIP := ctx.RealIP()

	// Check rate limit
	if err := service.MyService.AuthToken().CheckRateLimit(clientIP, service.TokenTypeMagicLink); err != nil {
		return ctx.JSON(common_err.TOO_MANY_REQUEST,
			model.Result{Success: common_err.TOO_MANY_REQUEST, Message: err.Error()})
	}

	// Create auth token
	_, plaintextToken, plaintextCode, err := service.MyService.AuthToken().CreateAuthToken(
		user.Id, userEmail, service.TokenTypeMagicLink, clientIP)
	if err != nil {
		logger.Error("Failed to create magic link token", zap.Error(err))
		return ctx.JSON(common_err.SERVICE_ERROR,
			model.Result{Success: common_err.SERVICE_ERROR, Message: "Failed to generate magic link"})
	}

	// Build magic link URL
	baseURL := getBaseURL(ctx)
	magicLink := fmt.Sprintf("%s/#/auth/magic?token=%s", baseURL, plaintextToken)

	// Generate email content
	emailContent := email.GenerateMagicLinkEmail(plaintextCode, magicLink)

	// Send email asynchronously (don't make user wait for SMTP)
	go func() {
		if err := sendEmailViaSMTP(userEmail, emailContent); err != nil {
			logger.Error("Failed to send magic link email", zap.Error(err))
		}
	}()

	// Return immediately to user (don't wait for email to send)
	return ctx.JSON(common_err.SUCCESS,
		model.Result{Success: common_err.SUCCESS, Message: "Magic link sent to your email"})
}

// PostMagicLinkVerify verifies a magic link token or code and logs the user in
// @Summary Verify magic link
// @Description Verify magic link token or code and return JWT
// @Tags auth
// @Accept json
// @Produce json
// @Param token body string false "Magic link token"
// @Param code body string false "6-digit code"
// @Param email body string false "Email (required for code)"
// @Success 200 {object} model.Result
// @Router /v1/auth/magic-link/verify [post]
func PostMagicLinkVerify(ctx echo.Context) error {
	json := make(map[string]string)
	ctx.Bind(&json)

	token := strings.TrimSpace(json["token"])
	code := strings.TrimSpace(json["code"])
	userEmail := strings.TrimSpace(json["email"])

	var authToken *model2.AuthTokenModel
	var err error

	// Verify either by token or code
	if token != "" {
		authToken, err = service.MyService.AuthToken().ValidateToken(token)
	} else if code != "" && userEmail != "" {
		authToken, err = service.MyService.AuthToken().ValidateCode(userEmail, code, service.TokenTypeMagicLink)
	} else {
		return ctx.JSON(common_err.CLIENT_ERROR,
			model.Result{Success: common_err.INVALID_PARAMS, Message: "Token or code (with email) required"})
	}

	if err != nil {
		return ctx.JSON(common_err.CLIENT_ERROR,
			model.Result{Success: common_err.INVALID_PARAMS, Message: "Invalid or expired magic link"})
	}

	// Consume token (one-time use)
	if err := service.MyService.AuthToken().ConsumeToken(authToken.ID); err != nil {
		logger.Error("Failed to consume magic link token", zap.Error(err))
	}

	// Get user
	user := service.MyService.User().GetUserAllInfoById(fmt.Sprintf("%d", authToken.UserID))
	if user.Id == 0 {
		return ctx.JSON(common_err.SERVICE_ERROR,
			model.Result{Success: common_err.USER_NOT_EXIST, Message: "User not found"})
	}

	// Generate JWT tokens
	privateKey, _ := service.MyService.User().GetKeyPair()

	accessToken, err := jwt.GetAccessToken(user.Username, privateKey, user.Id)
	if err != nil {
		return ctx.JSON(common_err.SERVICE_ERROR,
			model.Result{Success: common_err.SERVICE_ERROR, Message: err.Error()})
	}

	refreshToken, err := jwt.GetRefreshToken(user.Username, privateKey, user.Id)
	if err != nil {
		return ctx.JSON(common_err.SERVICE_ERROR,
			model.Result{Success: common_err.SERVICE_ERROR, Message: err.Error()})
	}

	tokenData := system_model.VerifyInformation{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(3 * time.Hour).Unix(),
	}

	data := make(map[string]interface{}, 2)
	user.Password = ""
	data["token"] = tokenData
	data["user"] = user

	logger.Info("Magic link login successful", zap.String("email", authToken.Email), zap.Int("user_id", user.Id))

	return ctx.JSON(common_err.SUCCESS,
		model.Result{Success: common_err.SUCCESS, Message: "Login successful", Data: data})
}

// EmailLoginEnabled reports whether passwordless email sign-in is available on
// this PCS. It requires USER_EMAIL plus the SMTP host/port to all be configured.
// GetUserStatus exposes this so the login UI only offers the email-link option
// when email can actually be sent.
func EmailLoginEnabled() bool {
	return os.Getenv("USER_EMAIL") != "" &&
		os.Getenv("SMTP_HOST") != "" &&
		os.Getenv("SMTP_PORT") != ""
}

// getUserByEmail looks up a user by checking the input email against the
// USER_EMAIL env var. If it matches, returns the first (only) user on this PCS.
func getUserByEmail(inputEmail string) model2.UserDBModel {
	configuredEmail := os.Getenv("USER_EMAIL")
	if configuredEmail == "" || !strings.EqualFold(inputEmail, configuredEmail) {
		return model2.UserDBModel{}
	}
	// Email matches — return the PCS user (single user system)
	users := service.MyService.User().GetAllUserName()
	if len(users) == 0 {
		return model2.UserDBModel{}
	}
	return service.MyService.User().GetUserAllInfoByName(users[0].Username)
}

// getBaseURL returns the base URL for email links using existing PCS env vars.
// REF_DOMAIN (e.g., "alice.nsl.sh") and REF_SCHEME (e.g., "https") are set
// in docker-compose.yml from the DOMAIN variable.
// Falls back to request headers if env vars are not set.
func getBaseURL(ctx echo.Context) string {
	domain := os.Getenv("REF_DOMAIN")
	scheme := os.Getenv("REF_SCHEME")
	if domain != "" && scheme != "" {
		return fmt.Sprintf("%s://%s", scheme, domain)
	}
	// Fallback: derive from request (may be wrong behind reverse proxies)
	return fmt.Sprintf("%s://%s", ctx.Scheme(), ctx.Request().Host)
}

// unencryptedAuth allows SMTP authentication over unencrypted connections
// This is needed because our internal SMTP relay doesn't use TLS
type unencryptedAuth struct {
	username, password string
}

func (a unencryptedAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "PLAIN", []byte("\x00" + a.username + "\x00" + a.password), nil
}

func (a unencryptedAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		return nil, fmt.Errorf("unexpected server challenge")
	}
	return nil, nil
}

// sendEmailViaSMTP sends an email via the local SMTP relay using the email builder
func sendEmailViaSMTP(to string, content email.EmailContent) error {
	// Get SMTP configuration
	smtpHost := os.Getenv("SMTP_HOST")
	if smtpHost == "" {
		smtpHost = "smtp" // Default to 'smtp' container hostname on pcs network
	}
	smtpPort := os.Getenv("SMTP_PORT")
	if smtpPort == "" {
		smtpPort = "587"
	}

	smtpUsername := os.Getenv("SMTP_USERNAME")
	smtpPassword := os.Getenv("SMTP_PASSWORD")

	if smtpUsername == "" || smtpPassword == "" {
		// Use default credentials
		smtpUsername = "casaos"
		smtpPassword = "localpassword"
	}

	// Build email message using the email builder (handles MIME with attachments)
	from := "auth@localhost"
	message, err := email.BuildMIMEEmail(from, to, content)
	if err != nil {
		logger.Error("Failed to build email", zap.Error(err))
		return fmt.Errorf("failed to build email: %w", err)
	}

	// Connect to SMTP server
	smtpAddr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)
	client, err := smtp.Dial(smtpAddr)
	if err != nil {
		logger.Error("Failed to connect to SMTP server", zap.String("address", smtpAddr), zap.Error(err))
		return fmt.Errorf("failed to connect to SMTP server %s: %w", smtpAddr, err)
	}
	defer client.Close()

	// Authenticate using custom auth that allows unencrypted connections for internal SMTP relay
	auth := unencryptedAuth{
		username: smtpUsername,
		password: smtpPassword,
	}
	if err = client.Auth(auth); err != nil {
		logger.Error("SMTP authentication failed", zap.Error(err))
		return fmt.Errorf("SMTP authentication failed: %w", err)
	}

	// Set sender and recipient
	if err = client.Mail(from); err != nil {
		logger.Error("Failed to set sender", zap.Error(err))
		return fmt.Errorf("failed to set sender: %w", err)
	}

	if err = client.Rcpt(to); err != nil {
		logger.Error("Failed to set recipient", zap.Error(err))
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	// Send message
	w, err := client.Data()
	if err != nil {
		logger.Error("Failed to send email data", zap.Error(err))
		return fmt.Errorf("failed to send email data: %w", err)
	}

	_, err = w.Write(message)
	if err != nil {
		logger.Error("Failed to write email", zap.Error(err))
		return fmt.Errorf("failed to write email: %w", err)
	}

	err = w.Close()
	if err != nil {
		logger.Error("Failed to send email", zap.Error(err))
		return fmt.Errorf("failed to send email: %w", err)
	}

	client.Quit()

	return nil
}
