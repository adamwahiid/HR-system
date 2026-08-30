package main

import "github.com/gin-gonic/gin"

func setupRouter() *gin.Engine {

	r := gin.Default()

	// Public routes — no token needed
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
		protected.PUT("/workers/:id", UpdateWorker)

		// Managers
		protected.GET("/managers", GetManagers)
		protected.GET("/managers/:id", GetManagerByID)
		protected.PUT("/managers/:id", UpdateManager)

		// Board Members
		protected.GET("/board-members", GetBoardMembers)
		protected.GET("/board-members/:id", GetBoardMemberByID)
		protected.PUT("/board-members/:id", UpdateBoardMember)
	}

	adminRes := r.Group("/")
	adminRes.Use(AuthMiddleware())
	adminRes.Use(RequireRole("admin", "hr"))
	{
		adminRes.DELETE("/workers/:id", DeleteWorker)
		adminRes.DELETE("/managers/:id", DeleteManager)
		adminRes.DELETE("/board-members/:id", DeleteBoardMember)
	}

	// Admin routes
	admin := r.Group("/admin")
	admin.Use(AuthMiddleware())
	admin.Use(RequireRole("admin", "hr"))
	{
		admin.POST("/users", CreateUser)
	}

	// Serve the frontend interface
	r.Static("/ui", "./public")

	return r
}
