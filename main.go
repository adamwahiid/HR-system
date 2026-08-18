package main

func main() {

	connectDB()

	r := setupRouter()

	r.Run(":8080")
}
