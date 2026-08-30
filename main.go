package main

import (
	"log"

	_ "github.com/joho/godotenv/autoload"
)

func main() {

	connectDB()

	// Perform database migration
	err := DB.AutoMigrate(&Role{}, &User{}, &Worker{}, &Manager{}, &BoardMember{})
	if err != nil {
		log.Fatalf("Failed to auto-migrate database: %v", err)
	}

	// Seed roles and initial admin account
	seedDatabase()

	r := setupRouter()

	r.Run(":8080")
}
