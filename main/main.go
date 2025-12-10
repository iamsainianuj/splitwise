package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"splitwise/main/internal/api"
	"splitwise/main/internal/db"
)

func main() {
	// Initialize database
	if err := db.Init(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Initialize API handler
	handler := api.NewHandler()

	// API Routes
	http.HandleFunc("/api/users", handler.EnableCORS(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.GetUsers(w, r)
		case http.MethodPost:
			handler.CreateUser(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	http.HandleFunc("/api/groups", handler.EnableCORS(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.GetGroups(w, r)
		case http.MethodPost:
			handler.CreateGroup(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	http.HandleFunc("/api/expenses", handler.EnableCORS(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.GetExpenses(w, r)
		case http.MethodPost:
			handler.AddExpense(w, r)
		case http.MethodDelete:
			handler.DeleteExpense(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	http.HandleFunc("/api/balances", handler.EnableCORS(handler.GetBalances))
	http.HandleFunc("/api/settle", handler.EnableCORS(handler.Settle))

	// Serve static files (web UI)
	fs := http.FileServer(http.Dir("./web"))
	http.Handle("/", fs)

	// Get port from environment variable (required for cloud hosting)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🚀 SplitWise server running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
