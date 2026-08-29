package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	RoleID   int    `json:"role_id" binding:"required"`
}

func Register(c *gin.Context) {

	var request RegisterRequest

	// Read JSON request
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Check if role exists
	var role Role

	if err := DB.First(&role, "role_id = ?", request.RoleID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid role_id",
		})
		return
	}

	// Hash password
	hashedPassword, err := HashPassword(request.Password)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to hash password",
		})
		return
	}

	// Create user
	user := User{
		Email:        request.Email,
		PasswordHash: hashedPassword,
		RoleID:       request.RoleID,
	}

	// Save user
	if err := DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Response
	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user_id": user.UserID,
		"email":   user.Email,
		"role_id": user.RoleID,
	})
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func Login(c *gin.Context) {

	var request LoginRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var user User

	// Find user by email
	if err := DB.Where("email = ?", request.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid email or password",
		})
		return
	}

	// Check password
	if !CheckPassword(request.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid email or password",
		})
		return
	}

	// Generate JWT
	token, err := GenerateToken(user)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not generate token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
		"user_id": user.UserID,
		"role_id": user.RoleID,
	})
}

func Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully. Please discard your token.",
	})
}