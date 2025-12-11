package db

import (
	"database/sql"
	"splitwise/main/internal/entity"
)

func CreateUser(userID, userName, email, passwordHash string) (*entity.User, error) {
	_, err := DB.Exec(
		"INSERT INTO users (user_id, user_name, user_email, password_hash) VALUES (?, ?, ?, ?)",
		userID, userName, email, passwordHash,
	)
	if err != nil {
		return nil, err
	}
	return &entity.User{UserID: userID, UserName: userName, UserEmail: email}, nil
}

func GetUserByEmail(email string) (*entity.User, string, error) {
	user := &entity.User{}
	var passwordHash string
	err := DB.QueryRow(
		"SELECT user_id, user_name, user_email, password_hash FROM users WHERE user_email = ?",
		email,
	).Scan(&user.UserID, &user.UserName, &user.UserEmail, &passwordHash)
	if err != nil {
		return nil, "", err
	}
	return user, passwordHash, nil
}

func GetAllUsers() ([]*entity.User, error) {
	rows, err := DB.Query("SELECT user_id, user_name, user_email FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]*entity.User, 0)
	for rows.Next() {
		user := &entity.User{}
		if err := rows.Scan(&user.UserID, &user.UserName, &user.UserEmail); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func GetUserByID(userID string) (*entity.User, error) {
	user := &entity.User{}
	err := DB.QueryRow(
		"SELECT user_id, user_name, user_email FROM users WHERE user_id = ?",
		userID,
	).Scan(&user.UserID, &user.UserName, &user.UserEmail)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func SearchUsers(query string, excludeUserID string) ([]*entity.User, error) {
	rows, err := DB.Query(`
		SELECT user_id, user_name, user_email FROM users 
		WHERE user_id != ? AND (user_name LIKE ? OR user_email LIKE ?)
		LIMIT 10
	`, excludeUserID, "%"+query+"%", "%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]*entity.User, 0)
	for rows.Next() {
		user := &entity.User{}
		if err := rows.Scan(&user.UserID, &user.UserName, &user.UserEmail); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func EmailExists(email string) bool {
	var count int
	DB.QueryRow("SELECT COUNT(*) FROM users WHERE user_email = ?", email).Scan(&count)
	return count > 0
}

// Session management in DB
func SaveSession(token, userID string, expiresAt string) error {
	_, err := DB.Exec(
		"INSERT OR REPLACE INTO sessions (token, user_id, expires_at) VALUES (?, ?, ?)",
		token, userID, expiresAt,
	)
	return err
}

func GetSession(token string) (string, error) {
	var userID string
	err := DB.QueryRow(
		"SELECT user_id FROM sessions WHERE token = ? AND expires_at > datetime('now')",
		token,
	).Scan(&userID)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return userID, err
}

func DeleteSession(token string) error {
	_, err := DB.Exec("DELETE FROM sessions WHERE token = ?", token)
	return err
}
