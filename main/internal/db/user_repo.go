package db

import "splitwise/main/internal/entity"

func CreateUser(user *entity.User) error {
	_, err := DB.Exec(
		"INSERT INTO users (user_id, user_name, user_email) VALUES (?, ?, ?)",
		user.UserID, user.UserName, user.UserEmail,
	)
	return err
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

