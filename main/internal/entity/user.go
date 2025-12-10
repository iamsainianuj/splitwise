package entity

type User struct {
	UserID    string `json:"user_id"`
	UserName  string `json:"user_name"`
	UserEmail string `json:"user_email"`
}

func NewUser(userID, userName, userEmail string) *User {
	return &User{
		UserID:    userID,
		UserName:  userName,
		UserEmail: userEmail,
	}
}
func (u *User) GetUserID() string {
	return u.UserID
}
func (u *User) GetUserName() string {
	return u.UserName
}
func (u *User) GetUserEmail() string {
	return u.UserEmail
}
