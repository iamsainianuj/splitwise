package entity

type Split struct {
	User   *User   `json:"user"`
	Amount float64 `json:"amount"`
}

func NewSplit(user *User, amount float64) *Split {
	return &Split{
		User:   user,
		Amount: amount,
	}
}
func (s *Split) GetUser() *User {
	return s.User
}
func (s *Split) GetAmount() float64 {
	return s.Amount
}
