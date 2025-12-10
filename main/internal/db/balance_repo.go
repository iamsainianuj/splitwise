package db

type BalanceRecord struct {
	FromUserID   string
	FromUserName string
	ToUserID     string
	ToUserName   string
	Amount       float64
}

func UpdateBalance(paidByUserID string, splitUserID string, amount float64) error {
	// First, try to update existing balance
	result, err := DB.Exec(`
		UPDATE balances 
		SET amount = amount + ? 
		WHERE from_user_id = ? AND to_user_id = ?
	`, amount, splitUserID, paidByUserID)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		// Insert new balance record
		_, err = DB.Exec(`
			INSERT INTO balances (from_user_id, to_user_id, amount) 
			VALUES (?, ?, ?)
		`, splitUserID, paidByUserID, amount)
		if err != nil {
			return err
		}
	}

	// Also update the reverse direction
	result, err = DB.Exec(`
		UPDATE balances 
		SET amount = amount - ? 
		WHERE from_user_id = ? AND to_user_id = ?
	`, amount, paidByUserID, splitUserID)
	if err != nil {
		return err
	}

	rowsAffected, _ = result.RowsAffected()
	if rowsAffected == 0 {
		_, err = DB.Exec(`
			INSERT INTO balances (from_user_id, to_user_id, amount) 
			VALUES (?, ?, ?)
		`, paidByUserID, splitUserID, -amount)
		if err != nil {
			return err
		}
	}

	return nil
}

func SettleBalance(fromUserID, toUserID string, amount float64) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update from -> to
	_, err = tx.Exec(`
		UPDATE balances 
		SET amount = amount - ? 
		WHERE from_user_id = ? AND to_user_id = ?
	`, amount, fromUserID, toUserID)
	if err != nil {
		return err
	}

	// Update to -> from
	_, err = tx.Exec(`
		UPDATE balances 
		SET amount = amount + ? 
		WHERE from_user_id = ? AND to_user_id = ?
	`, amount, toUserID, fromUserID)
	if err != nil {
		return err
	}

	// Clean up zero balances
	_, err = tx.Exec("DELETE FROM balances WHERE amount = 0")
	if err != nil {
		return err
	}

	return tx.Commit()
}

func GetAllBalances() ([]BalanceRecord, error) {
	rows, err := DB.Query(`
		SELECT b.from_user_id, u1.user_name, b.to_user_id, u2.user_name, b.amount
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
		if err := rows.Scan(&b.FromUserID, &b.FromUserName, &b.ToUserID, &b.ToUserName, &b.Amount); err != nil {
			return nil, err
		}
		balances = append(balances, b)
	}
	return balances, nil
}

