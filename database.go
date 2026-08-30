package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func connectDB() {

	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, password, dbname, port)

	var err error

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("Connected to PostgreSQL successfully!")
}

func seedDatabase() {
	// 1. Seed exact Roles
	roles := []Role{
		{RoleID: 1, RoleName: "HR"},
		{RoleID: 2, RoleName: "Admin"},
		{RoleID: 3, RoleName: "Worker"},
		{RoleID: 4, RoleName: "Manager"},
		{RoleID: 5, RoleName: "Board Member"},
	}

	for _, r := range roles {
		// Only create if it doesn't exist to prevent duplicates
		DB.FirstOrCreate(&r, Role{RoleID: r.RoleID})
	}

	// 2. Seed default Admin if the users table is completely empty
	var count int64
	DB.Model(&User{}).Count(&count)
	if count == 0 {
		hash, _ := HashPassword("password123")
		admin := User{
			Email:        "admin@company.com",
			PasswordHash: hash,
			RoleID:       2, // Admin role
		}
		DB.Create(&admin)
		fmt.Println("Seeded default admin account -> Email: admin@company.com | Password: password123")
	}
}
