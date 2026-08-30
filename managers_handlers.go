package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Helper function to populate names for a list of managers
func populateManagersNames(managers []Manager) {
	var boardMems []BoardMember
	if err := DB.Find(&boardMems).Error; err != nil {
		return
	}
	var workers []Worker
	if err := DB.Find(&workers).Error; err != nil {
		return
	}

	boardMap := make(map[int]string)
	for _, b := range boardMems {
		boardMap[b.BoardMemID] = b.Name
	}

	workersMap := make(map[int][]string)
	for _, w := range workers {
		workersMap[w.ManagerID] = append(workersMap[w.ManagerID], w.Name)
	}

	for i := range managers {
		managers[i].ManagerName = "None"
		if bName, ok := boardMap[managers[i].BoardMemID]; ok {
			managers[i].BoardMemName = bName
		} else {
			managers[i].BoardMemName = "None"
		}

		if wNames, ok := workersMap[managers[i].ManagerID]; ok {
			managers[i].WorkerNames = wNames
		} else {
			managers[i].WorkerNames = []string{}
		}
	}
}

// Helper function to populate names for a single manager
func populateManagerNames(manager *Manager) {
	manager.ManagerName = "None"
	var bm BoardMember
	if err := DB.First(&bm, "board_mem_id = ?", manager.BoardMemID).Error; err == nil {
		manager.BoardMemName = bm.Name
	} else {
		manager.BoardMemName = "None"
	}

	var workers []Worker
	if err := DB.Where("manager_id = ?", manager.ManagerID).Find(&workers).Error; err == nil {
		manager.WorkerNames = make([]string, len(workers))
		for i, w := range workers {
			manager.WorkerNames[i] = w.Name
		}
	} else {
		manager.WorkerNames = []string{}
	}
}

// Get all managers
func GetManagers(c *gin.Context) {
	roleName, _, err := GetUserRole(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var managers []Manager
	query := DB

	switch roleName {
	case "admin", "hr", "manager", "worker", "board member", "board_member":
		// Can see all
	default:
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	if err := query.Find(&managers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	populateManagersNames(managers)

	c.JSON(http.StatusOK, managers)
}

// Get manager by ID
func GetManagerByID(c *gin.Context) {
	id := c.Param("id")
	roleName, _, err := GetUserRole(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var manager Manager
	if err := DB.First(&manager, "manager_id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Manager not found"})
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

	populateManagerNames(&manager)

	c.JSON(http.StatusOK, manager)
}

type UpdateManagerRequest struct {
	Name       string   `json:"name"`
	Email      string   `json:"email"`
	Password   string   `json:"password"`
	Salary     *float64 `json:"salary"`
	BoardMemID *int     `json:"board_mem_id"`
}

// Update manager
func UpdateManager(c *gin.Context) {
	id := c.Param("id")
	roleName, loggedInUserID, err := GetUserRole(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var manager Manager
	if err := DB.First(&manager, "manager_id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Manager not found"})
		return
	}

	// Ownership check
	switch roleName {
	case "admin", "hr":
		// Allowed
	case "board member", "board_member":
		var bm BoardMember
		if err := DB.First(&bm, "user_id = ?", loggedInUserID).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Board Member profile not found"})
			return
		}
		if manager.BoardMemID != bm.BoardMemID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: Not your manager"})
			return
		}
	case "manager":
		if manager.UserID == nil || *manager.UserID != loggedInUserID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: Not your profile"})
			return
		}
	default:
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	var req UpdateManagerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields allowed for all
	if req.Name != "" {
		manager.Name = req.Name
	}
	if req.Email != "" {
		manager.Email = req.Email
	}

	// Update restricted fields only if admin/hr
	if roleName == "admin" || roleName == "hr" {
		if req.Salary != nil {
			manager.Salary = *req.Salary
		}
		if req.BoardMemID != nil {
			var checkBoardMem BoardMember
			if err := DB.First(&checkBoardMem, "board_mem_id = ?", *req.BoardMemID).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Provided board_mem_id does not exist"})
				return
			}
			manager.BoardMemID = *req.BoardMemID
		}
	}

	err = DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&manager).Error; err != nil {
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

		if len(updates) > 0 && manager.UserID != nil {
			if err := tx.Model(&User{}).Where("user_id = ?", *manager.UserID).Updates(updates).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	populateManagerNames(&manager)

	c.JSON(http.StatusOK, manager)
}

// Delete manager
func DeleteManager(c *gin.Context) {
	id := c.Param("id")

	var manager Manager
	if err := DB.First(&manager, "manager_id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Manager not found"})
		return
	}

	err := DB.Transaction(func(tx *gorm.DB) error {
		// Delete manager profile
		if err := tx.Delete(&manager).Error; err != nil {
			return err
		}

		// Delete user account
		if manager.UserID != nil {
			if err := tx.Delete(&User{}, "user_id = ?", *manager.UserID).Error; err != nil {
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
		"message": "Manager deleted successfully",
	})
}
