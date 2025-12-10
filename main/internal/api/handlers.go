package api

import (
	"encoding/json"
	"net/http"
	"splitwise/main/internal/db"
	"splitwise/main/internal/entity"
	"splitwise/main/internal/stragegy"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

// Request/Response types
type CreateUserRequest struct {
	UserID    string `json:"user_id"`
	UserName  string `json:"user_name"`
	UserEmail string `json:"user_email"`
}

type CreateGroupRequest struct {
	GroupID   string   `json:"group_id"`
	GroupName string   `json:"group_name"`
	MemberIDs []string `json:"member_ids"`
}

type AddExpenseRequest struct {
	ExpenseID          string             `json:"expense_id"`
	ExpenseDescription string             `json:"expense_description"`
	ExpenseAmount      float64            `json:"expense_amount"`
	PaidByUserID       string             `json:"paid_by_user_id"`
	GroupID            string             `json:"group_id"`
	SplitType          string             `json:"split_type"`
	SplitData          map[string]float64 `json:"split_data"`
}

type SettleRequest struct {
	FromUserID string  `json:"from_user_id"`
	ToUserID   string  `json:"to_user_id"`
	Amount     float64 `json:"amount"`
}

type BalanceEntry struct {
	UserID   string  `json:"user_id"`
	UserName string  `json:"user_name"`
	Amount   float64 `json:"amount"`
}

type UserBalances struct {
	UserID   string         `json:"user_id"`
	UserName string         `json:"user_name"`
	Balances []BalanceEntry `json:"balances"`
}

type ExpenseResponse struct {
	ExpenseID          string  `json:"expense_id"`
	ExpenseDescription string  `json:"expense_description"`
	ExpenseAmount      float64 `json:"expense_amount"`
	GroupID            string  `json:"group_id"`
	GroupName          string  `json:"group_name"`
	PaidByUserID       string  `json:"paid_by_user_id"`
	PaidByUserName     string  `json:"paid_by_user_name"`
}

// CORS middleware
func (h *Handler) EnableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user := entity.NewUser(req.UserID, req.UserName, req.UserEmail)
	if err := db.CreateUser(user); err != nil {
		http.Error(w, "Failed to create user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *Handler) GetUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	users, err := db.GetAllUsers()
	if err != nil {
		http.Error(w, "Failed to get users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (h *Handler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	members := make([]*entity.User, 0)
	for _, memberID := range req.MemberIDs {
		user, err := db.GetUserByID(memberID)
		if err == nil {
			members = append(members, user)
		}
	}

	group := entity.NewGroup(req.GroupID, req.GroupName, members)
	if err := db.CreateGroup(group); err != nil {
		http.Error(w, "Failed to create group: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(group)
}

func (h *Handler) GetGroups(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	groups, err := db.GetAllGroups()
	if err != nil {
		http.Error(w, "Failed to get groups", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(groups)
}

func (h *Handler) AddExpense(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AddExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get group for split calculation
	group, err := db.GetGroupByID(req.GroupID)
	if err != nil {
		http.Error(w, "Group not found", http.StatusBadRequest)
		return
	}

	// Convert split type string to enum
	var splitType stragegy.SplitType
	switch req.SplitType {
	case "equal":
		splitType = stragegy.Equal
	case "percentage":
		splitType = stragegy.Percentage
	case "exact":
		splitType = stragegy.Exact
	default:
		splitType = stragegy.Equal
	}

	// Convert split data from userID keys to User keys
	splitData := make(map[entity.User]float64)
	for userID, amount := range req.SplitData {
		user, err := db.GetUserByID(userID)
		if err == nil {
			splitData[*user] = amount
		}
	}

	// Calculate splits using strategy
	splits := stragegy.GetSplitStrategy(splitType, group).CalculateSplits(splitData, req.ExpenseAmount)

	// Save expense to database
	if err := db.CreateExpense(req.ExpenseID, req.ExpenseDescription, req.ExpenseAmount, req.GroupID, req.PaidByUserID, splits); err != nil {
		http.Error(w, "Failed to add expense: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Update balances in database
	for _, split := range splits {
		if split.User.UserID != req.PaidByUserID {
			if err := db.UpdateBalance(req.PaidByUserID, split.User.UserID, split.Amount); err != nil {
				// Log error but continue
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "created", "expense_id": req.ExpenseID})
}

func (h *Handler) GetExpenses(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	expenseRecords, err := db.GetAllExpenses()
	if err != nil {
		http.Error(w, "Failed to get expenses", http.StatusInternalServerError)
		return
	}

	expenses := make([]ExpenseResponse, 0)
	for _, exp := range expenseRecords {
		expenses = append(expenses, ExpenseResponse{
			ExpenseID:          exp.ExpenseID,
			ExpenseDescription: exp.ExpenseDescription,
			ExpenseAmount:      exp.ExpenseAmount,
			GroupID:            exp.GroupID,
			GroupName:          exp.GroupName,
			PaidByUserID:       exp.PaidByUserID,
			PaidByUserName:     exp.PaidByUserName,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(expenses)
}

func (h *Handler) Settle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SettleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := db.SettleBalance(req.FromUserID, req.ToUserID, req.Amount); err != nil {
		http.Error(w, "Failed to settle: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "settled"})
}

func (h *Handler) GetBalances(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	balanceRecords, err := db.GetAllBalances()
	if err != nil {
		http.Error(w, "Failed to get balances", http.StatusInternalServerError)
		return
	}

	// Group balances by user
	userBalanceMap := make(map[string]*UserBalances)
	for _, b := range balanceRecords {
		if _, exists := userBalanceMap[b.FromUserID]; !exists {
			userBalanceMap[b.FromUserID] = &UserBalances{
				UserID:   b.FromUserID,
				UserName: b.FromUserName,
				Balances: make([]BalanceEntry, 0),
			}
		}
		userBalanceMap[b.FromUserID].Balances = append(userBalanceMap[b.FromUserID].Balances, BalanceEntry{
			UserID:   b.ToUserID,
			UserName: b.ToUserName,
			Amount:   b.Amount,
		})
	}

	balances := make([]UserBalances, 0)
	for _, ub := range userBalanceMap {
		balances = append(balances, *ub)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(balances)
}

func (h *Handler) DeleteExpense(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	expenseID := r.URL.Query().Get("id")
	if expenseID == "" {
		http.Error(w, "Expense ID required", http.StatusBadRequest)
		return
	}

	if err := db.DeleteExpense(expenseID); err != nil {
		http.Error(w, "Failed to delete expense", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}
