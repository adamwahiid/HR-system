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

// CanViewSalary determines whether the currently authenticated user can view the salary of the target user profile.
// Admin and HR can view all salaries.
// Workers, managers, and board members can only view their own salary (when targetUserID matches loggedInUserID).
func CanViewSalary(roleName string, loggedInUserID int, targetUserID *int) bool {
	switch strings.ToLower(strings.TrimSpace(roleName)) {
	case "admin", "hr":
		return true
	default:
		return targetUserID != nil && *targetUserID == loggedInUserID
	}
}

// MaskWorkerSalary masks the worker's salary if the caller is not authorized to view it
func MaskWorkerSalary(w *Worker, roleName string, loggedInUserID int) {
	if !CanViewSalary(roleName, loggedInUserID, w.UserID) {
		w.Salary = nil
	}
}

// MaskWorkersSalaries masks the salaries of workers the caller is not authorized to view
func MaskWorkersSalaries(workers []Worker, roleName string, loggedInUserID int) {
	for i := range workers {
		MaskWorkerSalary(&workers[i], roleName, loggedInUserID)
	}
}

// MaskManagerSalary masks the manager's salary if the caller is not authorized to view it
func MaskManagerSalary(m *Manager, roleName string, loggedInUserID int) {
	if !CanViewSalary(roleName, loggedInUserID, m.UserID) {
		m.Salary = nil
	}
}

// MaskManagersSalaries masks the salaries of managers the caller is not authorized to view
func MaskManagersSalaries(managers []Manager, roleName string, loggedInUserID int) {
	for i := range managers {
		MaskManagerSalary(&managers[i], roleName, loggedInUserID)
	}
}

// MaskBoardMemberSalary masks the board member's salary if the caller is not authorized to view it
func MaskBoardMemberSalary(bm *BoardMember, roleName string, loggedInUserID int) {
	if !CanViewSalary(roleName, loggedInUserID, bm.UserID) {
		bm.Salary = nil
	}
}

// MaskBoardMembersSalaries masks the salaries of board members the caller is not authorized to view
func MaskBoardMembersSalaries(boardMembers []BoardMember, roleName string, loggedInUserID int) {
	for i := range boardMembers {
		MaskBoardMemberSalary(&boardMembers[i], roleName, loggedInUserID)
	}
}
