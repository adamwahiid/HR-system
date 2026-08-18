package main

import "github.com/gin-gonic/gin"

func setupRouter() *gin.Engine {

	r := gin.Default()

	// Authentication
	r.POST("/register", Register)

	// Workers
	r.GET("/workers", GetWorkers)
	r.GET("/workers/:id", GetWorkerByID)
	r.POST("/workers", CreateWorker)
	r.PUT("/workers/:id", UpdateWorker)
	r.DELETE("/workers/:id", DeleteWorker)

	// Managers
	r.GET("/managers", GetManagers)
	r.GET("/managers/:id", GetManagerByID)
	r.POST("/managers", CreateManager)
	r.PUT("/managers/:id", UpdateManager)
	r.DELETE("/managers/:id", DeleteManager)

	// Board Members
	r.GET("/board-members", GetBoardMembers)
	r.GET("/board-members/:id", GetBoardMemberByID)
	r.POST("/board-members", CreateBoardMember)
	r.PUT("/board-members/:id", UpdateBoardMember)
	r.DELETE("/board-members/:id", DeleteBoardMember)

	return r
}
