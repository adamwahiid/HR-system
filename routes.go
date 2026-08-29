package main

import "github.com/gin-gonic/gin"

func setupRouter() *gin.Engine {

	r := gin.Default()

	// Public routes — no token needed
	r.POST("/register", Register)
	r.POST("/login", Login)

	// Protected routes — token required
	protected := r.Group("/")
	protected.Use(AuthMiddleware())
	{
		//log out
		protected.POST("/logout", Logout)
		// Workers
		protected.GET("/workers", GetWorkers)
		protected.GET("/workers/:id", GetWorkerByID)
		protected.POST("/workers", CreateWorker)
		protected.PUT("/workers/:id", UpdateWorker)
		protected.DELETE("/workers/:id", DeleteWorker)

		// Managers
		protected.GET("/managers", GetManagers)
		protected.GET("/managers/:id", GetManagerByID)
		protected.POST("/managers", CreateManager)
		protected.PUT("/managers/:id", UpdateManager)
		protected.DELETE("/managers/:id", DeleteManager)

		// Board Members
		protected.GET("/board-members", GetBoardMembers)
		protected.GET("/board-members/:id", GetBoardMemberByID)
		protected.POST("/board-members", CreateBoardMember)
		protected.PUT("/board-members/:id", UpdateBoardMember)
		protected.DELETE("/board-members/:id", DeleteBoardMember)
	}

	return r
}