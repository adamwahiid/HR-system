package main

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreateUserRequest struct {
	Name       string  `json:"name" binding:"required"`
	Email      string  `json:"email" binding:"required,email"`
	Password   string  `json:"password" binding:"required"`
	RoleID     int     `json:"role_id" binding:"required"`
	Salary     float64 `json:"salary"`
	ManagerID  int     `json:"manager_id"`
	BoardMemID int     `json:"board_mem_id"`
}

func CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. Check if email already exists
	var existingUser User
	if err := DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// 2. Check if role_id exists
	var role Role
	if err := DB.First(&role, req.RoleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role_id"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	// 3. Hash the password
	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// 4. Execute transaction to create User and Profile
	var createdUser User
	err = DB.Transaction(func(tx *gorm.DB) error {
		// Create the login account in users
		createdUser = User{
			Email:        req.Email,
			PasswordHash: hashedPassword,
			RoleID:       req.RoleID,
		}

		if err := tx.Create(&createdUser).Error; err != nil {
			return err
		}

		// Based on role name, create respective profile
		switch role.RoleName {
		case "worker", "Worker":
			if req.ManagerID == 0 {
				return errors.New("manager_id is required for workers")
			}
			var checkManager Manager
			if err := tx.First(&checkManager, "manager_id = ?", req.ManagerID).Error; err != nil {
				return errors.New("provided manager_id does not exist")
			}
			worker := Worker{
				Name:      req.Name,
				Email:     req.Email,
				Salary:    &req.Salary,
				RoleID:    req.RoleID,
				ManagerID: req.ManagerID,
				UserID:    &createdUser.UserID,
			}
			if err := tx.Create(&worker).Error; err != nil {
				return err
			}
		case "manager", "Manager":
			if req.BoardMemID == 0 {
				return errors.New("board_mem_id is required for managers")
			}
			var checkBoardMem BoardMember
			if err := tx.First(&checkBoardMem, "board_mem_id = ?", req.BoardMemID).Error; err != nil {
				return errors.New("provided board_mem_id does not exist")
			}
			manager := Manager{
				Name:       req.Name,
				Email:      req.Email,
				Salary:     &req.Salary,
				RoleID:     req.RoleID,
				BoardMemID: req.BoardMemID,
				UserID:     &createdUser.UserID,
			}
			if err := tx.Create(&manager).Error; err != nil {
				return err
			}
		case "board_member", "Board Member", "board member", "Board_Member":
			boardMember := BoardMember{
				Name:   req.Name,
				Email:  req.Email,
				Salary: &req.Salary,
				RoleID: req.RoleID,
				UserID: &createdUser.UserID,
			}
			if err := tx.Create(&boardMember).Error; err != nil {
				return err
			}
		case "admin", "Admin", "hr", "HR", "hr_admin", "admin_hr":
			// Admins and HRs don't have secondary profile tables, so just return nil
			return nil
		default:
			return errors.New("unsupported role for user creation")
		}

		return nil
	})

	if err != nil {
		errStr := err.Error()
		if errStr == "unsupported role for user creation" ||
			errStr == "manager_id is required for workers" ||
			errStr == "provided manager_id does not exist" ||
			errStr == "board_mem_id is required for managers" ||
			errStr == "provided board_mem_id does not exist" {
			c.JSON(http.StatusBadRequest, gin.H{"error": errStr})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user account: " + errStr})
		}
		return
	}

	// 5. Return success response (exclude password data)
	c.JSON(http.StatusCreated, gin.H{
		"message": "account created successfully",
		"user": gin.H{
			"user_id": createdUser.UserID,
			"email":   createdUser.Email,
			"role_id": createdUser.RoleID,
		},
	})
}
