package db

import "splitwise/main/internal/entity"

type ExpenseRecord struct {
	ExpenseID          string        `json:"expense_id"`
	ExpenseDescription string        `json:"expense_description"`
	ExpenseAmount      float64       `json:"expense_amount"`
	GroupID            string        `json:"group_id"`
	GroupName          string        `json:"group_name"`
	PaidByUserID       string        `json:"paid_by_user_id"`
	PaidByUserName     string        `json:"paid_by_user_name"`
	Splits             []SplitRecord `json:"splits"`
}

type SplitRecord struct {
	UserID   string  `json:"user_id"`
	UserName string  `json:"user_name"`
	Amount   float64 `json:"amount"`
}

func CreateExpense(expenseID, description string, amount float64, groupID, paidByUserID string, splits []*entity.Split) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert expense
	_, err = tx.Exec(
		"INSERT INTO expenses (expense_id, expense_description, expense_amount, group_id, paid_by_user_id) VALUES (?, ?, ?, ?, ?)",
		expenseID, description, amount, groupID, paidByUserID,
	)
	if err != nil {
		return err
	}

	// Insert splits
	for _, split := range splits {
		_, err = tx.Exec(
			"INSERT INTO splits (expense_id, user_id, amount) VALUES (?, ?, ?)",
			expenseID, split.User.UserID, split.Amount,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func GetGroupExpenses(groupID string) ([]ExpenseRecord, error) {
	rows, err := DB.Query(`
		SELECT e.expense_id, e.expense_description, e.expense_amount, 
			   e.group_id, g.group_name, e.paid_by_user_id, u.user_name
		FROM expenses e
		JOIN groups g ON e.group_id = g.group_id
		JOIN users u ON e.paid_by_user_id = u.user_id
		WHERE e.group_id = ?
		ORDER BY e.date_created DESC
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	expenses := make([]ExpenseRecord, 0)
	for rows.Next() {
		exp := ExpenseRecord{}
		if err := rows.Scan(
			&exp.ExpenseID, &exp.ExpenseDescription, &exp.ExpenseAmount,
			&exp.GroupID, &exp.GroupName, &exp.PaidByUserID, &exp.PaidByUserName,
		); err != nil {
			return nil, err
		}
		// Fetch splits for this expense
		exp.Splits, _ = GetExpenseSplits(exp.ExpenseID)
		expenses = append(expenses, exp)
	}
	return expenses, nil
}

func GetExpenseSplits(expenseID string) ([]SplitRecord, error) {
	rows, err := DB.Query(`
		SELECT s.user_id, u.user_name, s.amount
		FROM splits s
		JOIN users u ON s.user_id = u.user_id
		WHERE s.expense_id = ?
	`, expenseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	splits := make([]SplitRecord, 0)
	for rows.Next() {
		split := SplitRecord{}
		if err := rows.Scan(&split.UserID, &split.UserName, &split.Amount); err != nil {
			return nil, err
		}
		splits = append(splits, split)
	}
	return splits, nil
}

func GetAllExpenses() ([]ExpenseRecord, error) {
	rows, err := DB.Query(`
		SELECT e.expense_id, e.expense_description, e.expense_amount, 
			   e.group_id, g.group_name, e.paid_by_user_id, u.user_name
		FROM expenses e
		JOIN groups g ON e.group_id = g.group_id
		JOIN users u ON e.paid_by_user_id = u.user_id
		ORDER BY e.date_created DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	expenses := make([]ExpenseRecord, 0)
	for rows.Next() {
		exp := ExpenseRecord{}
		if err := rows.Scan(
			&exp.ExpenseID, &exp.ExpenseDescription, &exp.ExpenseAmount,
			&exp.GroupID, &exp.GroupName, &exp.PaidByUserID, &exp.PaidByUserName,
		); err != nil {
			return nil, err
		}
		expenses = append(expenses, exp)
	}
	return expenses, nil
}

func DeleteExpense(expenseID string) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete splits first
	_, err = tx.Exec("DELETE FROM splits WHERE expense_id = ?", expenseID)
	if err != nil {
		return err
	}

	// Delete expense
	_, err = tx.Exec("DELETE FROM expenses WHERE expense_id = ?", expenseID)
	if err != nil {
		return err
	}

	return tx.Commit()
}
