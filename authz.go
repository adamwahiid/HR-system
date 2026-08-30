package main

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetUserRole retrieves the role name and user_id of the currently authenticated user
func GetUserRole(c *gin.Context) (string, int, error) {
	roleID, exists := c.Get("role_id")
	if !exists {
		return "", 0, errors.New("Role ID missing from token")
	}

	roleIDFloat, ok := roleID.(float64)
	if !ok {
		return "", 0, errors.New("Invalid Role ID type")
	}

	userID, exists := c.Get("user_id")
	if !exists {
		return "", 0, errors.New("User ID missing from token")
	}

	userIDFloat, ok := userID.(float64)
	if !ok {
		return "", 0, errors.New("Invalid User ID type")
	}

	var role Role
	if err := DB.First(&role, int(roleIDFloat)).Error; err != nil {
		return "", 0, errors.New("Role not found")
	}

	return strings.ToLower(role.RoleName), int(userIDFloat), nil
}
