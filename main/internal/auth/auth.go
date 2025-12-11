package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Database reference (set by main)
var DB *sql.DB

type Session struct {
	UserID    string
	UserName  string
	ExpiresAt time.Time
}

// SetDB sets the database reference for session storage
func SetDB(db *sql.DB) {
	DB = db
}

// HashPassword creates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPassword compares a password with its hash
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateToken creates a random session token
func GenerateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// GenerateUserID creates a unique user ID
func GenerateUserID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// CreateSession stores a new session in the database
func CreateSession(userID, userName string) (string, error) {
	token, err := GenerateToken()
	if err != nil {
		return "", err
	}

	expiresAt := time.Now().Add(24 * time.Hour * 7) // 7 days

	// Store in database
	_, err = DB.Exec(
		"INSERT OR REPLACE INTO sessions (token, user_id, user_name, expires_at) VALUES (?, ?, ?, ?)",
		token, userID, userName, expiresAt.Format(time.RFC3339),
	)
	if err != nil {
		return "", err
	}

	return token, nil
}

// GetSession retrieves a session by token from the database
func GetSession(token string) *Session {
	if DB == nil {
		return nil
	}

	var userID, userName, expiresAtStr string
	err := DB.QueryRow(
		"SELECT user_id, user_name, expires_at FROM sessions WHERE token = ?",
		token,
	).Scan(&userID, &userName, &expiresAtStr)

	if err != nil {
		return nil
	}

	expiresAt, err := time.Parse(time.RFC3339, expiresAtStr)
	if err != nil {
		return nil
	}

	if time.Now().After(expiresAt) {
		// Session expired, delete it
		DeleteSession(token)
		return nil
	}

	return &Session{
		UserID:    userID,
		UserName:  userName,
		ExpiresAt: expiresAt,
	}
}

// DeleteSession removes a session from the database
func DeleteSession(token string) {
	if DB != nil {
		DB.Exec("DELETE FROM sessions WHERE token = ?", token)
	}
}

// GetUserFromRequest extracts user info from request cookie
func GetUserFromRequest(r *http.Request) *Session {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return nil
	}
	return GetSession(cookie.Value)
}

// AuthMiddleware checks if user is authenticated
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := GetUserFromRequest(r)
		if session == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

// CleanupExpiredSessions removes expired sessions from the database
func CleanupExpiredSessions() {
	if DB != nil {
		DB.Exec("DELETE FROM sessions WHERE expires_at < ?", time.Now().Format(time.RFC3339))
	}
}
