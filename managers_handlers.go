package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Get all managers
func GetManagers(c *gin.Context) {
	var managers []Manager

	if err := DB.Find(&managers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, managers)
}

// Get manager by ID
func GetManagerByID(c *gin.Context) {
	id := c.Param("id")

	var manager Manager

	if err := DB.First(&manager, "manager_id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Manager not found"})
		return
	}

	c.JSON(http.StatusOK, manager)
}

// Create manager
func CreateManager(c *gin.Context) {
	var manager Manager

	if err := c.ShouldBindJSON(&manager); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := DB.Create(&manager).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, manager)
}

// Update manager
func UpdateManager(c *gin.Context) {
	id := c.Param("id")

	var manager Manager

	if err := DB.First(&manager, "manager_id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Manager not found"})
		return
	}

	if err := c.ShouldBindJSON(&manager); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := DB.Save(&manager).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, manager)
}

// Delete manager
func DeleteManager(c *gin.Context) {
	id := c.Param("id")

	result := DB.Delete(&Manager{}, "manager_id = ?", id)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Manager not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Manager deleted successfully",
	})
}
