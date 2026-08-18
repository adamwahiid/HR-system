package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Get all workers
func GetWorkers(c *gin.Context) {
	var workers []Worker

	if err := DB.Find(&workers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, workers)
}

// Get worker by ID
func GetWorkerByID(c *gin.Context) {
	id := c.Param("id")

	var worker Worker

	if err := DB.First(&worker, "worker_id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Worker not found"})
		return
	}

	c.JSON(http.StatusOK, worker)
}

// Create worker
func CreateWorker(c *gin.Context) {
	var worker Worker

	if err := c.ShouldBindJSON(&worker); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := DB.Create(&worker).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, worker)
}

// Update worker
func UpdateWorker(c *gin.Context) {
	id := c.Param("id")

	var worker Worker

	if err := DB.First(&worker, "worker_id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Worker not found"})
		return
	}

	if err := c.ShouldBindJSON(&worker); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := DB.Save(&worker).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, worker)
}

// Delete worker
func DeleteWorker(c *gin.Context) {
	id := c.Param("id")

	result := DB.Delete(&Worker{}, "worker_id = ?", id)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Worker not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Worker deleted successfully",
	})
}
