package handlers

import (
	"genrent/internal/middleware"
	"genrent/internal/models"
	"genrent/internal/services/email"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type RegisterRequest struct {
	Name     string      `json:"name" binding:"required"`
	Email    string      `json:"email" binding:"required,email"`
	Phone    string      `json:"phone" binding:"required"`
	Password string      `json:"password" binding:"required,min=6"`
	Role     models.Role `json:"role"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func Register(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": ValidationError(err), "errors": ValidationErrors(err)})
			return
		}

		var existing models.User
		if result := db.Where("email = ?", req.Email).First(&existing); result.Error == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
			return
		}

		hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
			return
		}

		role := req.Role
		if role == "" || (role != models.RoleCustomer && role != models.RoleVendor) {
			role = models.RoleCustomer
		}

		user := models.User{
			Name:     req.Name,
			Email:    req.Email,
			Phone:    req.Phone,
			Password: string(hashed),
			Role:     role,
		}

		if result := db.Create(&user); result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "registration successful",
			"user": gin.H{
				"id":    user.ID,
				"name":  user.Name,
				"email": user.Email,
				"role":  user.Role,
			},
		})
	}
}

func Login(db *gorm.DB, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": ValidationError(err), "errors": ValidationErrors(err)})
			return
		}

		var user models.User
		if result := db.Where("email = ?", req.Email).First(&user); result.Error != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}

		claims := &middleware.Claims{
			UserID: user.ID,
			Email:  user.Email,
			Role:   user.Role,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour * 7)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenStr, err := token.SignedString([]byte(jwtSecret))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token": tokenStr,
			"user": gin.H{
				"id":    user.ID,
				"name":  user.Name,
				"email": user.Email,
				"phone": user.Phone,
				"role":  user.Role,
			},
		})
	}
}

func GetProfile(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		var user models.User
		if result := db.First(&user, userID); result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"phone": user.Phone,
			"role":  user.Role,
		})
	}
}

// ForgotPasswordRequest is the request payload for forgot password
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ForgotPassword initiates the password reset process
func ForgotPassword(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ForgotPasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": ValidationError(err), "errors": ValidationErrors(err)})
			return
		}

		var user models.User
		if result := db.Where("email = ?", req.Email).First(&user); result.Error != nil {
			// Don't reveal if email exists or not - always return success
			c.JSON(http.StatusOK, gin.H{
				"message": "If an account exists with this email, a password reset link has been sent.",
			})
			return
		}

		// Generate a unique reset token
		token := uuid.New().String()
		expiresAt := time.Now().Add(1 * time.Hour) // Token expires in 1 hour

		// Create password reset record
		passwordReset := models.PasswordReset{
			UserID:    user.ID,
			Token:     token,
			ExpiresAt: expiresAt,
		}

		if result := db.Create(&passwordReset); result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create password reset token"})
			return
		}

		// Send password reset email
		log.Printf("[ForgotPassword] Attempting to send reset email to %s (EmailEnabled: %v)", user.Email, emailCfg.Enabled)
		data := email.EmailData{
			To:      user.Email,
			ToName:  user.Name,
			Subject: "Reset Your GenRent Password",
			Message: token, // Pass token in message field
		}

		email.Send(emailCfg, data, email.PasswordResetEmail(data))
		log.Printf("[ForgotPassword] Password reset email queued for %s", user.Email)

		c.JSON(http.StatusOK, gin.H{
			"message": "If an account exists with this email, a password reset link has been sent.",
		})
	}
}

// ResetPasswordRequest is the request payload for resetting password
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// ResetPassword handles the actual password reset with valid token
func ResetPassword(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ResetPasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": ValidationError(err), "errors": ValidationErrors(err)})
			return
		}

		var passwordReset models.PasswordReset
		if result := db.Where("token = ? AND used_at IS NULL AND expires_at > NOW()", req.Token).First(&passwordReset); result.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired reset token"})
			return
		}

		// Hash the new password
		hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
			return
		}

		// Update user password
		if result := db.Model(&models.User{}).Where("id = ?", passwordReset.UserID).Update("password", string(hashed)); result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update password"})
			return
		}

		// Mark token as used
		now := time.Now()
		db.Model(&passwordReset).Update("used_at", now)

		// Invalidate all other unused tokens for this user
		db.Model(&models.PasswordReset{}).
			Where("user_id = ? AND id != ? AND used_at IS NULL", passwordReset.UserID, passwordReset.ID).
			Update("used_at", now)

		c.JSON(http.StatusOK, gin.H{
			"message": "Password reset successfully. You can now login with your new password.",
		})
	}
}
