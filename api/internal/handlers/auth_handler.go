package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pquerna/otp/totp"
	qrcode "github.com/skip2/go-qrcode"
	"github.com/terraconsole/api/internal/config"
	"github.com/terraconsole/api/internal/middleware"
	"github.com/terraconsole/api/internal/models"
	"github.com/terraconsole/api/internal/services"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"encoding/base64"
)

type AuthHandler struct {
	db        *gorm.DB
	cfg       *config.Config
	encryptor *services.EncryptionService
}

func NewAuthHandler(db *gorm.DB, cfg *config.Config, enc *services.EncryptionService) *AuthHandler {
	return &AuthHandler{db: db, cfg: cfg, encryptor: enc}
}

type SignupRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	TOTPCode string `json:"totp_code,omitempty"`
}

type AuthResponse struct {
	Token     string      `json:"token"`
	ExpiresAt int64       `json:"expires_at"`
	User      models.User `json:"user"`
}

func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	if req.Email == "" || req.Password == "" || req.Username == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Email, username, and password are required"})
		return
	}

	if len(req.Password) < 8 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Password must be at least 8 characters"})
		return
	}

	// Check if user exists
	var existingUser models.User
	if err := h.db.Where("email = ? OR username = ?", req.Email, req.Username).First(&existingUser).Error; err == nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "User with this email or username already exists"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to hash password"})
		return
	}

	user := models.User{
		Email:    req.Email,
		Username: req.Username,
		Password: string(hashedPassword),
		FullName: req.FullName,
		IsActive: true,
	}

	if err := h.db.Create(&user).Error; err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create user"})
		return
	}

	token, expiresAt, err := h.generateToken(&user)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
		return
	}

	writeJSON(w, http.StatusCreated, AuthResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      user,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	var user models.User
	if err := h.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid email or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid email or password"})
		return
	}

	// Check MFA
	if user.MFAEnabled {
		if req.TOTPCode == "" {
			writeJSON(w, http.StatusPreconditionRequired, map[string]string{
				"error":    "MFA code required",
				"mfa_required": "true",
			})
			return
		}

		secret, err := h.encryptor.Decrypt(user.MFASecret)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to verify MFA"})
			return
		}

		if !totp.Validate(req.TOTPCode, secret) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid MFA code"})
			return
		}
	}

	// Update last login
	now := time.Now()
	h.db.Model(&user).Update("last_login_at", &now)

	token, expiresAt, err := h.generateToken(&user)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
		return
	}

	writeJSON(w, http.StatusOK, AuthResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      user,
	})
}

func (h *AuthHandler) SetupMFA(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	if user == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "TerraConsole",
		AccountName: user.Email,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to generate MFA secret"})
		return
	}

	// Encrypt and store the secret temporarily
	encSecret, err := h.encryptor.Encrypt(key.Secret())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to encrypt MFA secret"})
		return
	}

	h.db.Model(user).Update("mfa_secret", encSecret)

	// Generate QR code
	png, err := qrcode.Encode(key.URL(), qrcode.Medium, 256)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to generate QR code"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"secret":  key.Secret(),
		"qr_code": "data:image/png;base64," + base64.StdEncoding.EncodeToString(png),
		"url":     key.URL(),
	})
}

func (h *AuthHandler) VerifyMFA(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	if user == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}

	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	secret, err := h.encryptor.Decrypt(user.MFASecret)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "MFA not set up"})
		return
	}

	if !totp.Validate(req.Code, secret) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid MFA code"})
		return
	}

	h.db.Model(user).Update("mfa_enabled", true)

	writeJSON(w, http.StatusOK, map[string]string{"message": "MFA enabled successfully"})
}

func (h *AuthHandler) DisableMFA(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	if user == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}

	h.db.Model(user).Updates(map[string]interface{}{
		"mfa_enabled": false,
		"mfa_secret":  "",
	})

	writeJSON(w, http.StatusOK, map[string]string{"message": "MFA disabled successfully"})
}

func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	if user == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	if user == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}

	token, expiresAt, err := h.generateToken(user)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
		return
	}

	writeJSON(w, http.StatusOK, AuthResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      *user,
	})
}

func (h *AuthHandler) generateToken(user *models.User) (string, int64, error) {
	expiresAt := time.Now().Add(24 * time.Hour)

	claims := jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"exp":   expiresAt.Unix(),
		"iat":   time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(h.cfg.JWTSecret))
	if err != nil {
		return "", 0, err
	}

	return tokenString, expiresAt.Unix(), nil
}
