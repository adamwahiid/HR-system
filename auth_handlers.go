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