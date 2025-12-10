package expense

import (
	"splitwise/main/internal/entity"
	"time"
)

type Expense struct {
	ExpenseID          string          `json:"expense_id"`
	ExpenseDescription string          `json:"expense_description"`
	ExpenseAmount      float64         `json:"expense_amount"`
	Group              *entity.Group   `json:"group"`
	PaidBy             *entity.User    `json:"user"`
	Splits             []*entity.Split `json:"splits"`
	DateCreated        time.Time       `json:"date_created"`
}

func NewExpense(expenseID, expenseDescription string, expenseAmount float64, group *entity.Group, paidBy *entity.User, splits []*entity.Split) *Expense {
	return &Expense{
		ExpenseID:          expenseID,
		ExpenseDescription: expenseDescription,
		ExpenseAmount:      expenseAmount,
		Group:              group,
		PaidBy:             paidBy,
		Splits:             splits,
		DateCreated:        time.Now(),
	}
}
func (e *Expense) GetAmount() float64 {
	totalAmount := 0.0
	for _, split := range e.Splits {
		totalAmount += split.Amount
	}
	return totalAmount
}

func (e *Expense) GetPaidBy() *entity.User {
	return e.PaidBy
}
func (e *Expense) GetSplits() []*entity.Split {
	return e.Splits
}
func (e *Expense) GetDateCreated() time.Time {
	return e.DateCreated
}
func (e *Expense) GetGroup() *entity.Group {
	return e.Group
}
func (e *Expense) GetExpenseID() string {
	return e.ExpenseID
}
func (e *Expense) GetExpenseDescription() string {
	return e.ExpenseDescription
}
func (e *Expense) GetExpenseAmount() float64 {
	return e.ExpenseAmount
}
