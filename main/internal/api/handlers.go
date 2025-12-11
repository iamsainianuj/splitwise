package api

import (
	"encoding/json"
	"net/http"
	"splitwise/main/internal/auth"
	"splitwise/main/internal/db"
	"splitwise/main/internal/entity"
	"splitwise/main/internal/stragegy"
	"time"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

// Request/Response types
type RegisterRequest struct {
	UserName string `json:"user_name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message,omitempty"`
	UserID   string `json:"user_id,omitempty"`
	UserName string `json:"user_name,omitempty"`
	Email    string `json:"email,omitempty"`
	IsAdmin  bool   `json:"is_admin,omitempty"`
}

// Admin credentials (in production, use environment variables)
const (
	AdminUsername = "admin"
	AdminPassword = "admin"
)

type CreateGroupRequest struct {
	GroupName string   `json:"group_name"`
	MemberIDs []string `json:"member_ids"`
}

type AddExpenseRequest struct {
	ExpenseDescription string             `json:"expense_description"`
	ExpenseAmount      float64            `json:"expense_amount"`
	PaidByUserID       string             `json:"paid_by_user_id"`
	GroupID            string             `json:"group_id"`
	SplitType          string             `json:"split_type"`
	SplitData          map[string]float64 `json:"split_data"`
}

type SettleRequest struct {
	GroupID  string  `json:"group_id"`
	ToUserID string  `json:"to_user_id"`
	Amount   float64 `json:"amount"`
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

type BalanceResponse struct {
	GroupID      string  `json:"group_id"`
	FromUserID   string  `json:"from_user_id"`
	FromUserName string  `json:"from_user_name"`
	ToUserID     string  `json:"to_user_id"`
	ToUserName   string  `json:"to_user_name"`
	Amount       float64 `json:"amount"`
}

// CORS middleware
func (h *Handler) EnableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

// ============ AUTH ENDPOINTS ============

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, AuthResponse{Success: false, Message: "Invalid request"})
		return
	}

	// Validate input
	if req.UserName == "" || req.Email == "" || req.Password == "" {
		sendJSON(w, AuthResponse{Success: false, Message: "All fields are required"})
		return
	}

	if len(req.Password) < 6 {
		sendJSON(w, AuthResponse{Success: false, Message: "Password must be at least 6 characters"})
		return
	}

	// Check if email exists
	if db.EmailExists(req.Email) {
		sendJSON(w, AuthResponse{Success: false, Message: "Email already registered"})
		return
	}

	// Hash password
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		sendJSON(w, AuthResponse{Success: false, Message: "Server error"})
		return
	}

	// Create user
	userID := auth.GenerateUserID()
	user, err := db.CreateUser(userID, req.UserName, req.Email, passwordHash)
	if err != nil {
		sendJSON(w, AuthResponse{Success: false, Message: "Failed to create user"})
		return
	}

	// Create session
	token, _ := auth.CreateSession(user.UserID, user.UserName)
	setSessionCookie(w, token)

	sendJSON(w, AuthResponse{
		Success:  true,
		UserID:   user.UserID,
		UserName: user.UserName,
		Email:    user.UserEmail,
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, AuthResponse{Success: false, Message: "Invalid request"})
		return
	}

	// Get user by email
	user, passwordHash, err := db.GetUserByEmail(req.Email)
	if err != nil {
		sendJSON(w, AuthResponse{Success: false, Message: "Invalid email or password"})
		return
	}

	// Check password
	if !auth.CheckPassword(req.Password, passwordHash) {
		sendJSON(w, AuthResponse{Success: false, Message: "Invalid email or password"})
		return
	}

	// Create session
	token, _ := auth.CreateSession(user.UserID, user.UserName)
	setSessionCookie(w, token)

	sendJSON(w, AuthResponse{
		Success:  true,
		UserID:   user.UserID,
		UserName: user.UserName,
		Email:    user.UserEmail,
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err == nil {
		auth.DeleteSession(cookie.Value)
	}

	// Clear cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	sendJSON(w, AuthResponse{Success: true, Message: "Logged out"})
}

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	session := auth.GetUserFromRequest(r)
	if session == nil {
		sendJSON(w, AuthResponse{Success: false, Message: "Not authenticated"})
		return
	}

	user, err := db.GetUserByID(session.UserID)
	if err != nil {
		sendJSON(w, AuthResponse{Success: false, Message: "User not found"})
		return
	}

	sendJSON(w, AuthResponse{
		Success:  true,
		UserID:   user.UserID,
		UserName: user.UserName,
		Email:    user.UserEmail,
	})
}

// ============ USER ENDPOINTS ============

func (h *Handler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	session := auth.GetUserFromRequest(r)
	if session == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		sendJSON(w, []*entity.User{})
		return
	}

	users, err := db.SearchUsers(query, session.UserID)
	if err != nil {
		sendJSON(w, []*entity.User{})
		return
	}

	sendJSON(w, users)
}

// ============ GROUP ENDPOINTS ============

func (h *Handler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	session := auth.GetUserFromRequest(r)
	if session == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Build member list (always include creator)
	members := make([]*entity.User, 0)

	// Add creator
	creator, _ := db.GetUserByID(session.UserID)
	if creator != nil {
		members = append(members, creator)
	}

	// Add other members
	for _, memberID := range req.MemberIDs {
		if memberID != session.UserID {
			user, err := db.GetUserByID(memberID)
			if err == nil {
				members = append(members, user)
			}
		}
	}

	groupID := auth.GenerateUserID()
	group := entity.NewGroup(groupID, req.GroupName, members)

	if err := db.CreateGroup(group, session.UserID); err != nil {
		http.Error(w, "Failed to create group: "+err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSON(w, group)
}

func (h *Handler) GetMyGroups(w http.ResponseWriter, r *http.Request) {
	session := auth.GetUserFromRequest(r)
	if session == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	groups, err := db.GetUserGroupsWithBalances(session.UserID)
	if err != nil {
		sendJSON(w, []db.GroupWithBalance{})
		return
	}

	sendJSON(w, groups)
}

type AddMemberRequest struct {
	GroupID string `json:"group_id"`
	UserID  string `json:"user_id"`
}

func (h *Handler) AddMemberToGroup(w http.ResponseWriter, r *http.Request) {
	session := auth.GetUserFromRequest(r)
	if session == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req AddMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Verify current user is in group
	if !db.IsUserInGroup(session.UserID, req.GroupID) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Add the new member
	if err := db.AddMemberToGroup(req.GroupID, req.UserID); err != nil {
		http.Error(w, "Failed to add member", http.StatusInternalServerError)
		return
	}

	sendJSON(w, map[string]string{"status": "added"})
}

func (h *Handler) GetGroupDetails(w http.ResponseWriter, r *http.Request) {
	session := auth.GetUserFromRequest(r)
	if session == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	groupID := r.URL.Query().Get("id")
	if groupID == "" {
		http.Error(w, "Group ID required", http.StatusBadRequest)
		return
	}

	// Verify user is in group
	if !db.IsUserInGroup(session.UserID, groupID) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	group, err := db.GetGroupByID(groupID)
	if err != nil {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	// Get group expenses and balances
	expenses, _ := db.GetGroupExpenses(groupID)
	balances, _ := db.GetGroupBalances(groupID)

	sendJSON(w, map[string]interface{}{
		"group":    group,
		"expenses": expenses,
		"balances": balances,
	})
}

// ============ EXPENSE ENDPOINTS ============

func (h *Handler) AddExpense(w http.ResponseWriter, r *http.Request) {
	session := auth.GetUserFromRequest(r)
	if session == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req AddExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Verify user is in group
	if !db.IsUserInGroup(session.UserID, req.GroupID) {
		http.Error(w, "Access denied", http.StatusForbidden)
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

	// Determine who paid (use request value or fall back to session user)
	paidByUserID := req.PaidByUserID
	if paidByUserID == "" {
		paidByUserID = session.UserID
	}

	// Save expense to database
	expenseID := auth.GenerateUserID()
	if err := db.CreateExpense(expenseID, req.ExpenseDescription, req.ExpenseAmount, req.GroupID, paidByUserID, splits); err != nil {
		http.Error(w, "Failed to add expense: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Update balances in database (per group)
	for _, split := range splits {
		if split.User.UserID != paidByUserID {
			db.UpdateBalance(req.GroupID, paidByUserID, split.User.UserID, split.Amount)
		}
	}

	sendJSON(w, map[string]string{"status": "created", "expense_id": expenseID})
}

func (h *Handler) GetGroupExpenses(w http.ResponseWriter, r *http.Request) {
	session := auth.GetUserFromRequest(r)
	if session == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	groupID := r.URL.Query().Get("group_id")
	if groupID == "" {
		http.Error(w, "Group ID required", http.StatusBadRequest)
		return
	}

	// Verify user is in group
	if !db.IsUserInGroup(session.UserID, groupID) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	expenses, err := db.GetGroupExpenses(groupID)
	if err != nil {
		sendJSON(w, []ExpenseResponse{})
		return
	}

	sendJSON(w, expenses)
}

// ============ BALANCE/SETTLE ENDPOINTS ============

func (h *Handler) GetMyBalanceSummary(w http.ResponseWriter, r *http.Request) {
	session := auth.GetUserFromRequest(r)
	if session == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	summary, err := db.GetUserBalanceSummary(session.UserID)
	if err != nil {
		http.Error(w, "Failed to get balance summary", http.StatusInternalServerError)
		return
	}

	sendJSON(w, summary)
}

func (h *Handler) GetGroupBalances(w http.ResponseWriter, r *http.Request) {
	session := auth.GetUserFromRequest(r)
	if session == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	groupID := r.URL.Query().Get("group_id")
	if groupID == "" {
		http.Error(w, "Group ID required", http.StatusBadRequest)
		return
	}

	// Verify user is in group
	if !db.IsUserInGroup(session.UserID, groupID) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	balances, err := db.GetGroupBalances(groupID)
	if err != nil {
		sendJSON(w, []BalanceResponse{})
		return
	}

	response := make([]BalanceResponse, 0)
	for _, b := range balances {
		response = append(response, BalanceResponse{
			GroupID:      b.GroupID,
			FromUserID:   b.FromUserID,
			FromUserName: b.FromUserName,
			ToUserID:     b.ToUserID,
			ToUserName:   b.ToUserName,
			Amount:       b.Amount,
		})
	}

	sendJSON(w, response)
}

func (h *Handler) Settle(w http.ResponseWriter, r *http.Request) {
	session := auth.GetUserFromRequest(r)
	if session == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req SettleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Verify user is in group
	if !db.IsUserInGroup(session.UserID, req.GroupID) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	if err := db.SettleBalance(req.GroupID, session.UserID, req.ToUserID, req.Amount); err != nil {
		http.Error(w, "Failed to settle: "+err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSON(w, map[string]string{"status": "settled"})
}

// ============ ADMIN ENDPOINTS ============

func (h *Handler) AdminLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, AuthResponse{Success: false, Message: "Invalid request"})
		return
	}

	if req.Username != AdminUsername || req.Password != AdminPassword {
		sendJSON(w, AuthResponse{Success: false, Message: "Invalid credentials"})
		return
	}

	// Create admin session
	token, _ := auth.CreateSession("admin", "Administrator")
	setSessionCookie(w, token)

	sendJSON(w, AuthResponse{
		Success:  true,
		UserID:   "admin",
		UserName: "Administrator",
		IsAdmin:  true,
	})
}

func (h *Handler) isAdmin(r *http.Request) bool {
	session := auth.GetUserFromRequest(r)
	return session != nil && session.UserID == "admin"
}

func (h *Handler) AdminGetUsers(w http.ResponseWriter, r *http.Request) {
	if !h.isAdmin(r) {
		http.Error(w, "Admin access required", http.StatusForbidden)
		return
	}

	users, err := db.GetAllUsers()
	if err != nil {
		http.Error(w, "Failed to get users", http.StatusInternalServerError)
		return
	}

	sendJSON(w, users)
}

func (h *Handler) AdminGetGroups(w http.ResponseWriter, r *http.Request) {
	if !h.isAdmin(r) {
		http.Error(w, "Admin access required", http.StatusForbidden)
		return
	}

	groups, err := db.GetAllGroups()
	if err != nil {
		http.Error(w, "Failed to get groups", http.StatusInternalServerError)
		return
	}

	sendJSON(w, groups)
}

func (h *Handler) AdminDeleteUser(w http.ResponseWriter, r *http.Request) {
	if !h.isAdmin(r) {
		http.Error(w, "Admin access required", http.StatusForbidden)
		return
	}

	userID := r.URL.Query().Get("id")
	if userID == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	// Check if user has any pending balances (owes or is owed)
	hasBalance, balanceMsg, err := db.UserHasPendingBalances(userID)
	if err != nil {
		http.Error(w, "Failed to check balances", http.StatusInternalServerError)
		return
	}
	if hasBalance {
		sendJSON(w, map[string]interface{}{
			"success": false,
			"message": balanceMsg,
		})
		return
	}

	// Delete user
	if err := db.DeleteUser(userID); err != nil {
		http.Error(w, "Failed to delete user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSON(w, map[string]interface{}{
		"success": true,
		"message": "User deleted",
	})
}

func (h *Handler) AdminDeleteGroup(w http.ResponseWriter, r *http.Request) {
	if !h.isAdmin(r) {
		http.Error(w, "Admin access required", http.StatusForbidden)
		return
	}

	groupID := r.URL.Query().Get("id")
	if groupID == "" {
		http.Error(w, "Group ID required", http.StatusBadRequest)
		return
	}

	// Check if group has any unsettled balances
	hasBalance, err := db.GroupHasUnsettledBalances(groupID)
	if err != nil {
		http.Error(w, "Failed to check balances", http.StatusInternalServerError)
		return
	}
	if hasBalance {
		sendJSON(w, map[string]interface{}{
			"success": false,
			"message": "Cannot delete group with unsettled balances. All members must settle up first.",
		})
		return
	}

	// Delete group
	if err := db.DeleteGroup(groupID); err != nil {
		http.Error(w, "Failed to delete group: "+err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSON(w, map[string]interface{}{
		"success": true,
		"message": "Group deleted",
	})
}

// ============ HELPERS ============

func sendJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}
