package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Helper function to populate names for a list of board members
func populateBoardMembersNames(boardMems []BoardMember) {
	var managers []Manager
	if err := DB.Find(&managers).Error; err != nil {
		return
	}
	var workers []Worker
	if err := DB.Find(&workers).Error; err != nil {
		return
	}

	managersMap := make(map[int][]string)
	managerToBoardMap := make(map[int]int)
	for _, m := range managers {
		managersMap[m.BoardMemID] = append(managersMap[m.BoardMemID], m.Name)
		managerToBoardMap[m.ManagerID] = m.BoardMemID
	}

	workersMap := make(map[int][]string)
	for _, w := range workers {
		if boardMemID, ok := managerToBoardMap[w.ManagerID]; ok {
			workersMap[boardMemID] = append(workersMap[boardMemID], w.Name)
		}
	}

	for i := range boardMems {
		boardMems[i].ManagerName = "None"
		boardMems[i].BoardMemName = boardMems[i].Name

		if mNames, ok := managersMap[boardMems[i].BoardMemID]; ok {
			boardMems[i].ManagerNames = mNames
		} else {
			boardMems[i].ManagerNames = []string{}
		}

		if wNames, ok := workersMap[boardMems[i].BoardMemID]; ok {
			boardMems[i].WorkerNames = wNames
		} else {
			boardMems[i].WorkerNames = []string{}
		}
	}
}

// Helper function to populate names for a single board member
func populateBoardMemberNames(bm *BoardMember) {
	bm.ManagerName = "None"
	bm.BoardMemName = bm.Name

	var managers []Manager
	if err := DB.Where("board_mem_id = ?", bm.BoardMemID).Find(&managers).Error; err == nil {
		bm.ManagerNames = make([]string, len(managers))
		managerIDs := make([]int, len(managers))
		for i, m := range managers {
			bm.ManagerNames[i] = m.Name
			managerIDs[i] = m.ManagerID
		}

		if len(managerIDs) > 0 {
			var workers []Worker
			if err := DB.Where("manager_id IN ?", managerIDs).Find(&workers).Error; err == nil {
				bm.WorkerNames = make([]string, len(workers))
				for i, w := range workers {
					bm.WorkerNames[i] = w.Name
				}
			} else {
				bm.WorkerNames = []string{}
			}
		} else {
			bm.WorkerNames = []string{}
		}
	} else {
		bm.ManagerNames = []string{}
		bm.WorkerNames = []string{}
	}
}

// Get all board members
func GetBoardMembers(c *gin.Context) {
	roleName, loggedInUserID, err := GetUserRole(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var boardMembers []BoardMember
	query := DB

	switch roleName {
	case "admin", "hr", "manager", "worker", "board member", "board_member":
		// Can see all
	default:
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	if err := query.Find(&boardMembers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	populateBoardMembersNames(boardMembers)
	MaskBoardMembersSalaries(boardMembers, roleName, loggedInUserID)

	c.JSON(http.StatusOK, boardMembers)
}

// Get board member by ID
func GetBoardMemberByID(c *gin.Context) {
	id := c.Param("id")
	roleName, loggedInUserID, err := GetUserRole(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var boardMember BoardMember
	if err := DB.First(&boardMember, "board_mem_id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Board member not found"})
		return
	}

	// Ownership check
	switch roleName {
	case "admin", "hr", "manager", "worker", "board member", "board_member":
		// Allowed
	default:
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	populateBoardMemberNames(&boardMember)
	MaskBoardMemberSalary(&boardMember, roleName, loggedInUserID)

	c.JSON(http.StatusOK, boardMember)
}

type UpdateBoardMemberRequest struct {
	Name     string   `json:"name"`
	Email    string   `json:"email"`
	Password string   `json:"password"`
	Salary   *float64 `json:"salary"`
}

// Update board member
func UpdateBoardMember(c *gin.Context) {
	id := c.Param("id")
	roleName, loggedInUserID, err := GetUserRole(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var boardMember BoardMember
	if err := DB.First(&boardMember, "board_mem_id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Board member not found"})
		return
	}

	// Ownership check
	switch roleName {
	case "admin", "hr":
		// Allowed
	case "board member", "board_member":
		if boardMember.UserID == nil || *boardMember.UserID != loggedInUserID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: Not your profile"})
			return
		}
	default:
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	var req UpdateBoardMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields allowed for all
	if req.Name != "" {
		boardMember.Name = req.Name
	}
	if req.Email != "" {
		boardMember.Email = req.Email
	}

	// Update restricted fields only if admin/hr
	if roleName == "admin" || roleName == "hr" {
		if req.Salary != nil {
			boardMember.Salary = req.Salary
		}
	}

	err = DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&boardMember).Error; err != nil {
			return err
		}

		updates := map[string]interface{}{}
		if req.Email != "" {
			updates["email"] = req.Email
		}
		if req.Password != "" {
			hash, hashErr := HashPassword(req.Password)
			if hashErr != nil {
				return hashErr
			}
			updates["password_hash"] = hash
		}

		if len(updates) > 0 && boardMember.UserID != nil {
			if err := tx.Model(&User{}).Where("user_id = ?", *boardMember.UserID).Updates(updates).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	populateBoardMemberNames(&boardMember)
	MaskBoardMemberSalary(&boardMember, roleName, loggedInUserID)

	c.JSON(http.StatusOK, boardMember)
}

// Delete board member
func DeleteBoardMember(c *gin.Context) {
	id := c.Param("id")

	var boardMember BoardMember
	if err := DB.First(&boardMember, "board_mem_id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Board member not found"})
		return
	}

	err := DB.Transaction(func(tx *gorm.DB) error {
		// Delete profile
		if err := tx.Delete(&boardMember).Error; err != nil {
			return err
		}

		// Delete user account
		if boardMember.UserID != nil {
			if err := tx.Delete(&User{}, "user_id = ?", *boardMember.UserID).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Board member deleted successfully",
	})
}
