package handlers

import (
	"database/sql"
	"net/http"
	"simpleRegistration/internal/models"
	"simpleRegistration/internal/usecase"
	"time"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	//db        *config.Database
	jwtSecret []byte
	// Add token expiration configuration
	tokenExpiration time.Duration
	userUsecase     *usecase.UserUsecase
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(jwtSecret []byte, userUsecase *usecase.UserUsecase) *AuthHandler {
	return &AuthHandler{
		//db:              db,
		jwtSecret:       jwtSecret,
		tokenExpiration: 24 * time.Hour, // Default 24 hour expiration
		userUsecase:     userUsecase,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var user models.UserRegister

	// Validate input JSON
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid input format",
			"details": err.Error(),
		})
		return
	}

	// Check if user already exists
	exists, err := h.userUsecase.CheckUsernameExistence(user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	newUser := models.UserRegister{Username: user.Username, Password: user.Password}
	username, err := h.userUsecase.CreateUser(newUser)
	if err != nil {
		//tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User creation failed"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "User registered successfully",
		"username": username,
	})
}

// Login handles user authentication and JWT generation
func (h *AuthHandler) Login(c *gin.Context) {
	var login models.UserLogin
	if err := c.ShouldBindJSON(&login); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid login data"})
		return
	}

	// Get user from database
	var user models.User
	user, err := h.userUsecase.Find(login.Username)

	if err == sql.ErrNoRows {
		// Don't specify whether email or password was wrong
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Login process failed"})
		return
	}

	// Verify password
	if !h.userUsecase.CheckPassword(login.Password, user.Password) {
		// Use same message as above for security
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT with claims
	tokenString, err := h.userUsecase.GenerateJWT(user, h.tokenExpiration, h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
		return
	}

	// Return token with expiration
	c.JSON(http.StatusOK, gin.H{
		"token":      tokenString,
		"expires_in": h.tokenExpiration.Seconds(),
		"token_type": "Bearer",
	})
}

func (h *AuthHandler) Hello(c *gin.Context) {
	// Retrieve the username from the context
	username, exists := c.Get("username")
	if !exists {
		c.JSON(401, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Respond with a hello message
	c.JSON(200, gin.H{
		"message": "Hello " + username.(string),
	})
}
