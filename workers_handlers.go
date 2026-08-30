package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Helper function to populate names for a list of workers
func populateWorkersNames(workers []Worker) {
	var managers []Manager
	if err := DB.Find(&managers).Error; err != nil {
		return
	}
	var boardMems []BoardMember
	if err := DB.Find(&boardMems).Error; err != nil {
		return
	}

	managerMap := make(map[int]struct {
		Name       string
		BoardMemID int
	})
	for _, m := range managers {
		managerMap[m.ManagerID] = struct {
			Name       string
			BoardMemID int
		}{m.Name, m.BoardMemID}
	}

	boardMap := make(map[int]string)
	for _, b := range boardMems {
		boardMap[b.BoardMemID] = b.Name
	}

	for i := range workers {
		if mInfo, ok := managerMap[workers[i].ManagerID]; ok {
			workers[i].ManagerName = mInfo.Name
			if bName, ok2 := boardMap[mInfo.BoardMemID]; ok2 {
				workers[i].BoardMemName = bName
			} else {
				workers[i].BoardMemName = "None"
			}
		} else {
			workers[i].ManagerName = "None"
			workers[i].BoardMemName = "None"
		}
	}
}

// Helper function to populate names for a single worker
func populateWorkerNames(worker *Worker) {
	var manager Manager
	if err := DB.First(&manager, "manager_id = ?", worker.ManagerID).Error; err == nil {
		worker.ManagerName = manager.Name
		var bm BoardMember
		if err := DB.First(&bm, "board_mem_id = ?", manager.BoardMemID).Error; err == nil {
			worker.BoardMemName = bm.Name
		} else {
			worker.BoardMemName = "None"
		}
	} else {
		worker.ManagerName = "None"
		worker.BoardMemName = "None"
	}
}

// Get all workers
func GetWorkers(c *gin.Context) {
	roleName, _, err := GetUserRole(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var workers []Worker
	query := DB

	switch roleName {
	case "admin", "hr", "manager", "worker", "board member", "board_member":
		// Can see all
	default:
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	if err := query.Find(&workers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	populateWorkersNames(workers)

	c.JSON(http.StatusOK, workers)
}

// Get worker by ID
func GetWorkerByID(c *gin.Context) {
	id := c.Param("id")
	roleName, _, err := GetUserRole(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var worker Worker
	if err := DB.First(&worker, "worker_id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Worker not found"})
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

	populateWorkerNames(&worker)

	c.JSON(http.StatusOK, worker)
}

type UpdateWorkerRequest struct {
	Name      string   `json:"name"`
	Email     string   `json:"email"`
	Password  string   `json:"password"`
	Salary    *float64 `json:"salary"`
	ManagerID *int     `json:"manager_id"`
}

// Update worker
func UpdateWorker(c *gin.Context) {
	id := c.Param("id")
	roleName, loggedInUserID, err := GetUserRole(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var worker Worker
	if err := DB.First(&worker, "worker_id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Worker not found"})
		return
	}

	// Ownership check
	switch roleName {
	case "admin", "hr":
		// Allowed
	case "manager":
		var manager Manager
		if err := DB.First(&manager, "user_id = ?", loggedInUserID).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Manager profile not found"})
			return
		}
		if worker.ManagerID != manager.ManagerID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: Not your worker"})
			return
		}
	case "worker":
		if worker.UserID == nil || *worker.UserID != loggedInUserID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: Not your profile"})
			return
		}
	default:
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	var req UpdateWorkerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields allowed for all
	if req.Name != "" {
		worker.Name = req.Name
	}
	if req.Email != "" {
		worker.Email = req.Email
	}

	// Update restricted fields only if admin/hr
	if roleName == "admin" || roleName == "hr" {
		if req.Salary != nil {
			worker.Salary = *req.Salary
		}
		if req.ManagerID != nil {
			var checkManager Manager
			if err := DB.First(&checkManager, "manager_id = ?", *req.ManagerID).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Provided manager_id does not exist"})
				return
			}
			worker.ManagerID = *req.ManagerID
		}
	}

	err = DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&worker).Error; err != nil {
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

		if len(updates) > 0 && worker.UserID != nil {
			if err := tx.Model(&User{}).Where("user_id = ?", *worker.UserID).Updates(updates).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	populateWorkerNames(&worker)

	c.JSON(http.StatusOK, worker)
}

// Delete worker
func DeleteWorker(c *gin.Context) {
	id := c.Param("id")

	var worker Worker
	if err := DB.First(&worker, "worker_id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Worker not found"})
		return
	}

	err := DB.Transaction(func(tx *gorm.DB) error {
		// Delete worker profile
		if err := tx.Delete(&worker).Error; err != nil {
			return err
		}

		// Delete user account
		if worker.UserID != nil {
			if err := tx.Delete(&User{}, "user_id = ?", *worker.UserID).Error; err != nil {
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
		"message": "Worker deleted successfully",
	})
}
