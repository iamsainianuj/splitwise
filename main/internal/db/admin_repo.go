package db

import "fmt"

// UserHasPendingBalances checks if a user owes or is owed money
func UserHasPendingBalances(userID string) (bool, string, error) {
	// Check if user owes anyone
	var owesCount int
	var owesAmount float64
	err := DB.QueryRow(`
		SELECT COUNT(*), COALESCE(SUM(amount), 0) 
		FROM balances 
		WHERE from_user_id = ? AND amount > 0
	`, userID).Scan(&owesCount, &owesAmount)
	if err != nil {
		return false, "", err
	}
	if owesCount > 0 {
		return true, fmt.Sprintf("User owes $%.2f to others. Must settle up before deletion.", owesAmount), nil
	}

	// Check if anyone owes this user
	var owedCount int
	var owedAmount float64
	err = DB.QueryRow(`
		SELECT COUNT(*), COALESCE(SUM(amount), 0) 
		FROM balances 
		WHERE to_user_id = ? AND amount > 0
	`, userID).Scan(&owedCount, &owedAmount)
	if err != nil {
		return false, "", err
	}
	if owedCount > 0 {
		return true, fmt.Sprintf("User is owed $%.2f by others. Must collect before deletion.", owedAmount), nil
	}

	return false, "", nil
}

// GroupHasUnsettledBalances checks if a group has any non-zero balances
func GroupHasUnsettledBalances(groupID string) (bool, error) {
	var count int
	err := DB.QueryRow(`
		SELECT COUNT(*) FROM balances 
		WHERE group_id = ? AND amount != 0
	`, groupID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// DeleteUser removes a user and their data
func DeleteUser(userID string) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Remove from group memberships
	_, err = tx.Exec("DELETE FROM group_members WHERE user_id = ?", userID)
	if err != nil {
		return err
	}

	// Delete user's sessions
	_, err = tx.Exec("DELETE FROM sessions WHERE user_id = ?", userID)
	if err != nil {
		return err
	}

	// Delete the user
	_, err = tx.Exec("DELETE FROM users WHERE user_id = ?", userID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// DeleteGroup removes a group and all its data
func DeleteGroup(groupID string) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete splits for expenses in this group
	_, err = tx.Exec(`
		DELETE FROM splits WHERE expense_id IN 
		(SELECT expense_id FROM expenses WHERE group_id = ?)
	`, groupID)
	if err != nil {
		return err
	}

	// Delete expenses
	_, err = tx.Exec("DELETE FROM expenses WHERE group_id = ?", groupID)
	if err != nil {
		return err
	}

	// Delete balances
	_, err = tx.Exec("DELETE FROM balances WHERE group_id = ?", groupID)
	if err != nil {
		return err
	}

	// Delete group members
	_, err = tx.Exec("DELETE FROM group_members WHERE group_id = ?", groupID)
	if err != nil {
		return err
	}

	// Delete the group
	_, err = tx.Exec("DELETE FROM groups WHERE group_id = ?", groupID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

