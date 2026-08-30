# Employee Management System - Full Source Code

This document contains the complete source code for the Go backend. You can export this document to a PDF by right-clicking in your Markdown viewer/editor (like VS Code) and selecting **Export to PDF**, or by printing this page from your browser and selecting **Save as PDF**.

---

## 1. `main.go`
```go
package main

func main() {

	connectDB()

	r := setupRouter()

	r.Run(":8080")
}
```

---

## 2. `database.go`
```go
package main

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func connectDB() {

	dsn := "host=localhost user=postgres password=1234 dbname=project_1 port=5432 sslmode=disable"

	var err error

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("Connected to PostgreSQL successfully!")
}
```

---

## 3. `models.go`
```go
package main

type Role struct {
	RoleID   int    `gorm:"column:role_id;primaryKey;autoIncrement" json:"role_id"`
	RoleName string `gorm:"column:role_name;unique;not null" json:"role_name"`
}

func (Role) TableName() string {
	return "roles"
}

type User struct {
	UserID       int    `gorm:"column:user_id;primaryKey;autoIncrement" json:"user_id"`
	Email        string `gorm:"column:email;unique;not null" json:"email"`
	PasswordHash string `gorm:"column:password_hash;not null" json:"-"`
	RoleID       int    `gorm:"column:role_id;not null" json:"role_id"`
}

func (User) TableName() string {
	return "users"
}

type Worker struct {
	WorkerID     int     `gorm:"column:worker_id;primaryKey;autoIncrement" json:"worker_id"`
	Name         string  `gorm:"column:name;not null" json:"name"`
	Email        string  `gorm:"column:email;unique;not null" json:"email"`
	Salary       float64 `gorm:"column:salary;not null" json:"salary"`
	RoleID       int     `gorm:"column:role_id;not null" json:"role_id"`
	ManagerID    int     `gorm:"column:manager_id;not null" json:"manager_id"`
	UserID       *int    `gorm:"column:user_id" json:"user_id"`
	ManagerName  string  `gorm:"-" json:"manager_name"`
	BoardMemName string  `gorm:"-" json:"board_mem_name"`
}

func (Worker) TableName() string {
	return "workers"
}

type Manager struct {
	ManagerID    int      `gorm:"column:manager_id;primaryKey;autoIncrement" json:"manager_id"`
	Name         string   `gorm:"column:name;not null" json:"name"`
	Email        string   `gorm:"column:email;unique;not null" json:"email"`
	Salary       float64  `gorm:"column:salary;not null" json:"salary"`
	RoleID       int      `gorm:"column:role_id;not null" json:"role_id"`
	BoardMemID   int      `gorm:"column:board_mem_id" json:"board_mem_id"`
	UserID       *int     `gorm:"column:user_id" json:"user_id"`
	ManagerName  string   `gorm:"-" json:"manager_name"`
	BoardMemName string   `gorm:"-" json:"board_mem_name"`
	WorkerNames  []string `gorm:"-" json:"workers"`
}

func (Manager) TableName() string {
	return "managers"
}

type BoardMember struct {
	BoardMemID   int      `gorm:"column:board_mem_id;primaryKey;autoIncrement" json:"board_mem_id"`
	Name         string   `gorm:"column:name;not null" json:"name"`
	Email        string   `gorm:"column:email;unique;not null" json:"email"`
	Salary       float64  `gorm:"column:salary;not null" json:"salary"`
	RoleID       int      `gorm:"column:role_id;not null" json:"role_id"`
	UserID       *int     `gorm:"column:user_id" json:"user_id"`
	ManagerName  string   `gorm:"-" json:"manager_name"`
	BoardMemName string   `gorm:"-" json:"board_mem_name"`
	ManagerNames []string `gorm:"-" json:"managers"`
	WorkerNames  []string `gorm:"-" json:"workers"`
}

func (BoardMember) TableName() string {
	return "board_members"
}
```

---

## 4. `auth.go`
```go
package main

import (
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret []byte

func init() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}
	jwtSecret = []byte(secret)
}

func HashPassword(password string) (string, error) {

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)

	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func CheckPassword(password string, hash string) bool {

	err := bcrypt.CompareHashAndPassword(
		[]byte(hash),
		[]byte(password),
	)

	return err == nil
}

func GenerateToken(user User) (string, error) {

	claims := jwt.MapClaims{
		"user_id": user.UserID,
		"role_id": user.RoleID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	return token.SignedString(jwtSecret)
}
```

---

## 5. `authz.go`
```go
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
```

---

## 6. `middleware.go`
```go
package main

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header"})
			c.Abort()
			return
		}

		// Expecting format: "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		claims := jwt.MapClaims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Attach user info to context so handlers can access it later
		c.Set("user_id", claims["user_id"])
		c.Set("role_id", claims["role_id"])

		c.Next()
	}
}

// RequireRole checks if the user has one of the allowed roles
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleName, _, err := GetUserRole(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		for _, allowedRole := range allowedRoles {
			if roleName == strings.ToLower(allowedRole) {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: insufficient permissions"})
		c.Abort()
	}
}
```

---

## 7. `routes.go`
```go
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

	return r
}
```

---

## 8. `auth_handlers.go`
```go
package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func Login(c *gin.Context) {

	var request LoginRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var user User

	// Find user by email
	if err := DB.Where("email = ?", request.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid email or password",
		})
		return
	}

	// Check password
	if !CheckPassword(request.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid email or password",
		})
		return
	}

	// Generate JWT
	token, err := GenerateToken(user)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not generate token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
		"user_id": user.UserID,
		"role_id": user.RoleID,
	})
}

func Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully. Please discard your token.",
	})
}
```

---

## 9. `admin_handlers.go`
```go
package main

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreateUserRequest struct {
	Name       string  `json:"name" binding:"required"`
	Email      string  `json:"email" binding:"required,email"`
	Password   string  `json:"password" binding:"required"`
	RoleID     int     `json:"role_id" binding:"required"`
	Salary     float64 `json:"salary"`
	ManagerID  int     `json:"manager_id"`
	BoardMemID int     `json:"board_mem_id"`
}

func CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. Check if email already exists
	var existingUser User
	if err := DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// 2. Check if role_id exists
	var role Role
	if err := DB.First(&role, req.RoleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role_id"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	// 3. Hash the password
	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// 4. Execute transaction to create User and Profile
	var createdUser User
	err = DB.Transaction(func(tx *gorm.DB) error {
		// Create the login account in users
		createdUser = User{
			Email:        req.Email,
			PasswordHash: hashedPassword,
			RoleID:       req.RoleID,
		}

		if err := tx.Create(&createdUser).Error; err != nil {
			return err
		}

		// Based on role name, create respective profile
		switch role.RoleName {
		case "worker", "Worker":
			if req.ManagerID == 0 {
				return errors.New("manager_id is required for workers")
			}
			var checkManager Manager
			if err := tx.First(&checkManager, "manager_id = ?", req.ManagerID).Error; err != nil {
				return errors.New("provided manager_id does not exist")
			}
			worker := Worker{
				Name:      req.Name,
				Email:     req.Email,
				Salary:    req.Salary,
				RoleID:    req.RoleID,
				ManagerID: req.ManagerID,
				UserID:    &createdUser.UserID,
			}
			if err := tx.Create(&worker).Error; err != nil {
				return err
			}
		case "manager", "Manager":
			if req.BoardMemID == 0 {
				return errors.New("board_mem_id is required for managers")
			}
			var checkBoardMem BoardMember
			if err := tx.First(&checkBoardMem, "board_mem_id = ?", req.BoardMemID).Error; err != nil {
				return errors.New("provided board_mem_id does not exist")
			}
			manager := Manager{
				Name:       req.Name,
				Email:      req.Email,
				Salary:     req.Salary,
				RoleID:     req.RoleID,
				BoardMemID: req.BoardMemID,
				UserID:     &createdUser.UserID,
			}
			if err := tx.Create(&manager).Error; err != nil {
				return err
			}
		case "board_member", "Board Member", "board member", "Board_Member":
			boardMember := BoardMember{
				Name:   req.Name,
				Email:  req.Email,
				Salary: req.Salary,
				RoleID: req.RoleID,
				UserID: &createdUser.UserID,
			}
			if err := tx.Create(&boardMember).Error; err != nil {
				return err
			}
		default:
			return errors.New("unsupported role for user creation")
		}

		return nil
	})

	if err != nil {
		errStr := err.Error()
		if errStr == "unsupported role for user creation" ||
			errStr == "manager_id is required for workers" ||
			errStr == "provided manager_id does not exist" ||
			errStr == "board_mem_id is required for managers" ||
			errStr == "provided board_mem_id does not exist" {
			c.JSON(http.StatusBadRequest, gin.H{"error": errStr})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user account: " + errStr})
		}
		return
	}

	// 5. Return success response (exclude password data)
	c.JSON(http.StatusCreated, gin.H{
		"message": "account created successfully",
		"user": gin.H{
			"user_id": createdUser.UserID,
			"email":   createdUser.Email,
			"role_id": createdUser.RoleID,
		},
	})
}
```

---

## 10. `workers_handlers.go`
```go
package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

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

	c.JSON(http.StatusOK, worker)
}

type UpdateWorkerRequest struct {
	Name      string   `json:"name"`
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

	// Update restricted fields only if admin/hr
	if roleName == "admin" || roleName == "hr" {
		if req.Salary != nil {
			worker.Salary = *req.Salary
		}
		if req.ManagerID != nil {
			worker.ManagerID = *req.ManagerID
		}
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
```

---

## 11. `managers_handlers.go`
```go
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
```

---

## 12. `board_mem_handlers.go`
```go
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
	roleName, _, err := GetUserRole(c)
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

	c.JSON(http.StatusOK, boardMembers)
}

// Get board member by ID
func GetBoardMemberByID(c *gin.Context) {
	id := c.Param("id")
	roleName, _, err := GetUserRole(c)
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
			boardMember.Salary = *req.Salary
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
```

