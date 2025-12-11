package db

type BalanceRecord struct {
	GroupID      string  `json:"group_id"`
	FromUserID   string  `json:"from_user_id"`
	FromUserName string  `json:"from_user_name"`
	ToUserID     string  `json:"to_user_id"`
	ToUserName   string  `json:"to_user_name"`
	Amount       float64 `json:"amount"`
}

func UpdateBalance(groupID, paidByUserID, splitUserID string, amount float64) error {
	if paidByUserID == splitUserID {
		return nil
	}

	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update: splitUser owes paidBy
	_, err = tx.Exec(`
		INSERT INTO balances (group_id, from_user_id, to_user_id, amount) 
		VALUES (?, ?, ?, ?)
		ON CONFLICT(group_id, from_user_id, to_user_id) 
		DO UPDATE SET amount = amount + ?
	`, groupID, splitUserID, paidByUserID, amount, amount)
	if err != nil {
		return err
	}

	// Update reverse: paidBy is owed by splitUser (negative)
	_, err = tx.Exec(`
		INSERT INTO balances (group_id, from_user_id, to_user_id, amount) 
		VALUES (?, ?, ?, ?)
		ON CONFLICT(group_id, from_user_id, to_user_id) 
		DO UPDATE SET amount = amount + ?
	`, groupID, paidByUserID, splitUserID, -amount, -amount)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func SettleBalance(groupID, fromUserID, toUserID string, amount float64) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Reduce what fromUser owes toUser
	_, err = tx.Exec(`
		UPDATE balances 
		SET amount = amount - ? 
		WHERE group_id = ? AND from_user_id = ? AND to_user_id = ?
	`, amount, groupID, fromUserID, toUserID)
	if err != nil {
		return err
	}

	// Reduce what toUser is owed by fromUser
	_, err = tx.Exec(`
		UPDATE balances 
		SET amount = amount + ? 
		WHERE group_id = ? AND from_user_id = ? AND to_user_id = ?
	`, amount, groupID, toUserID, fromUserID)
	if err != nil {
		return err
	}

	// Clean up zero/tiny balances
	_, err = tx.Exec("DELETE FROM balances WHERE ABS(amount) < 0.10")
	if err != nil {
		return err
	}

	return tx.Commit()
}

func GetGroupBalances(groupID string) ([]BalanceRecord, error) {
	rows, err := DB.Query(`
		SELECT b.group_id, b.from_user_id, u1.user_name, b.to_user_id, u2.user_name, b.amount
		FROM balances b
		JOIN users u1 ON b.from_user_id = u1.user_id
		JOIN users u2 ON b.to_user_id = u2.user_id
		WHERE b.group_id = ? AND b.amount > 0.10
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	balances := make([]BalanceRecord, 0)
	for rows.Next() {
		b := BalanceRecord{}
		if err := rows.Scan(&b.GroupID, &b.FromUserID, &b.FromUserName, &b.ToUserID, &b.ToUserName, &b.Amount); err != nil {
			return nil, err
		}
		balances = append(balances, b)
	}
	return balances, nil
}

func GetUserBalances(userID string) ([]BalanceRecord, error) {
	rows, err := DB.Query(`
		SELECT b.group_id, b.from_user_id, u1.user_name, b.to_user_id, u2.user_name, b.amount
		FROM balances b
		JOIN users u1 ON b.from_user_id = u1.user_id
		JOIN users u2 ON b.to_user_id = u2.user_id
		WHERE (b.from_user_id = ? OR b.to_user_id = ?) AND b.amount != 0
	`, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	balances := make([]BalanceRecord, 0)
	for rows.Next() {
		b := BalanceRecord{}
		if err := rows.Scan(&b.GroupID, &b.FromUserID, &b.FromUserName, &b.ToUserID, &b.ToUserName, &b.Amount); err != nil {
			return nil, err
		}
		balances = append(balances, b)
	}
	return balances, nil
}

type BalanceSummary struct {
	TotalYouOwe   float64 `json:"total_you_owe"`
	TotalOwedToYou float64 `json:"total_owed_to_you"`
	NetBalance    float64 `json:"net_balance"`
}

func GetUserBalanceSummary(userID string) (*BalanceSummary, error) {
	summary := &BalanceSummary{}

	// Total you owe to others
	err := DB.QueryRow(`
		SELECT COALESCE(SUM(amount), 0) 
		FROM balances 
		WHERE from_user_id = ? AND amount > 0.10
	`, userID).Scan(&summary.TotalYouOwe)
	if err != nil {
		return nil, err
	}

	// Total others owe you
	err = DB.QueryRow(`
		SELECT COALESCE(SUM(amount), 0) 
		FROM balances 
		WHERE to_user_id = ? AND amount > 0.10
	`, userID).Scan(&summary.TotalOwedToYou)
	if err != nil {
		return nil, err
	}

	// Net balance (positive = you're owed, negative = you owe)
	summary.NetBalance = summary.TotalOwedToYou - summary.TotalYouOwe

	return summary, nil
}

func GetAllBalances() ([]BalanceRecord, error) {
	rows, err := DB.Query(`
		SELECT b.group_id, b.from_user_id, u1.user_name, b.to_user_id, u2.user_name, b.amount
		FROM balances b
		JOIN users u1 ON b.from_user_id = u1.user_id
		JOIN users u2 ON b.to_user_id = u2.user_id
		WHERE b.amount != 0
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	balances := make([]BalanceRecord, 0)
	for rows.Next() {
		b := BalanceRecord{}
		if err := rows.Scan(&b.GroupID, &b.FromUserID, &b.FromUserName, &b.ToUserID, &b.ToUserName, &b.Amount); err != nil {
			return nil, err
		}
		balances = append(balances, b)
	}
	return balances, nil
}
