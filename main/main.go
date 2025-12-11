package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"splitwise/main/internal/api"
	"splitwise/main/internal/auth"
	"splitwise/main/internal/db"
)

func main() {
	// Initialize database
	if err := db.Init(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Set database for auth package (persistent sessions)
	auth.SetDB(db.DB)

	// Initialize API handler
	handler := api.NewHandler()

	// Auth routes (public)
	http.HandleFunc("/api/auth/register", handler.EnableCORS(handler.Register))
	http.HandleFunc("/api/auth/login", handler.EnableCORS(handler.Login))
	http.HandleFunc("/api/auth/logout", handler.EnableCORS(handler.Logout))
	http.HandleFunc("/api/auth/me", handler.EnableCORS(handler.GetMe))

	// User routes (protected)
	http.HandleFunc("/api/users/search", handler.EnableCORS(handler.SearchUsers))

	// Group routes (protected)
	http.HandleFunc("/api/groups", handler.EnableCORS(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.GetMyGroups(w, r)
		case http.MethodPost:
			handler.CreateGroup(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	http.HandleFunc("/api/groups/details", handler.EnableCORS(handler.GetGroupDetails))
	http.HandleFunc("/api/groups/add-member", handler.EnableCORS(handler.AddMemberToGroup))

	// Expense routes (protected)
	http.HandleFunc("/api/expenses", handler.EnableCORS(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.GetGroupExpenses(w, r)
		case http.MethodPost:
			handler.AddExpense(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// Balance routes (protected)
	http.HandleFunc("/api/balances", handler.EnableCORS(handler.GetGroupBalances))
	http.HandleFunc("/api/balances/summary", handler.EnableCORS(handler.GetMyBalanceSummary))
	http.HandleFunc("/api/settle", handler.EnableCORS(handler.Settle))

	// Admin routes
	http.HandleFunc("/api/admin/login", handler.EnableCORS(handler.AdminLogin))
	http.HandleFunc("/api/admin/users", handler.EnableCORS(handler.AdminGetUsers))
	http.HandleFunc("/api/admin/groups", handler.EnableCORS(handler.AdminGetGroups))
	http.HandleFunc("/api/admin/users/delete", handler.EnableCORS(handler.AdminDeleteUser))
	http.HandleFunc("/api/admin/groups/delete", handler.EnableCORS(handler.AdminDeleteGroup))

	// Serve static files (web UI)
	// Try multiple paths to find the web directory
	webDir := "./web"
	if _, err := os.Stat(webDir); os.IsNotExist(err) {
		// Try relative to executable
		exe, _ := os.Executable()
		webDir = filepath.Join(filepath.Dir(exe), "..", "web")
	}
	if _, err := os.Stat(webDir); os.IsNotExist(err) {
		// Try parent directory
		webDir = "../web"
	}
	log.Printf("Serving static files from: %s", webDir)
	fs := http.FileServer(http.Dir(webDir))
	http.Handle("/", fs)

	// Get port from environment variable
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🚀 SplitWise server running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
