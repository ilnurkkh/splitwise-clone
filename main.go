package main

import (
	"log"
	"net/http"
	"os"

	"splitwise-clone/db"
	"splitwise-clone/handlers"
	"splitwise-clone/middleware"

	"github.com/joho/godotenv"
)

func main() {
	// 1. Load environment variables (DATABASE_URL, JWT_SECRET, PORT)
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: No .env file found, using system variables")
	}

	// 2. Initialize Postgres Connection Pool
	db.ConnectDB()
	defer db.Pool.Close()

	// 3. Setup Go 1.22 Router
	mux := http.NewServeMux()

	// --- Public Routes (No authentication needed) ---
	mux.HandleFunc("POST /api/users/register", handlers.RegisterUser)
	mux.HandleFunc("POST /api/users/login", handlers.LoginUser)

	// --- Protected Application Routes (Requires valid JWT Token) ---
	mux.HandleFunc("POST /api/expenses", middleware.RequireAuth(handlers.CreateExpense))
	mux.HandleFunc("POST /api/friends", middleware.RequireAuth(handlers.AddFriend))
	mux.HandleFunc("POST /api/groups", middleware.RequireAuth(handlers.CreateGroup))
	mux.HandleFunc("GET /api/dashboard/balances", middleware.RequireAuth(handlers.GetDashboardBalances))
	mux.HandleFunc("GET /api/users/me", middleware.RequireAuth(handlers.GetUserProfile))

	// 4. Start the Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
