package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Get all board members
func GetBoardMembers(c *gin.Context) {
	var boardMembers []BoardMember

	if err := DB.Find(&boardMembers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, boardMembers)
}

// Get board member by ID
func GetBoardMemberByID(c *gin.Context) {
	id := c.Param("id")

	var boardMember BoardMember

	if err := DB.First(&boardMember, "board_mem_id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Board member not found"})
		return
	}

	c.JSON(http.StatusOK, boardMember)
}

// Create board member
func CreateBoardMember(c *gin.Context) {
	var boardMember BoardMember

	if err := c.ShouldBindJSON(&boardMember); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := DB.Create(&boardMember).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, boardMember)
}

// Update board member
func UpdateBoardMember(c *gin.Context) {
	id := c.Param("id")

	var boardMember BoardMember

	if err := DB.First(&boardMember, "board_mem_id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Board member not found"})
		return
	}

	if err := c.ShouldBindJSON(&boardMember); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := DB.Save(&boardMember).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, boardMember)
}

// Delete board member
func DeleteBoardMember(c *gin.Context) {
	id := c.Param("id")

	result := DB.Delete(&BoardMember{}, "board_mem_id = ?", id)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Board member not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Board member deleted successfully",
	})
}
