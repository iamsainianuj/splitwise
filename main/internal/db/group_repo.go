package db

import (
	"splitwise/main/internal/entity"
)

type GroupWithBalance struct {
	*entity.Group
	YouOwe      float64 `json:"you_owe"`
	YouAreOwed  float64 `json:"you_are_owed"`
	HasPending  bool    `json:"has_pending"`
}

func CreateGroup(group *entity.Group, createdBy string) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert group
	_, err = tx.Exec(
		"INSERT INTO groups (group_id, group_name, created_by) VALUES (?, ?, ?)",
		group.GroupID, group.GroupName, createdBy,
	)
	if err != nil {
		return err
	}

	// Insert group members (including creator)
	for _, member := range group.GroupMembers {
		_, err = tx.Exec(
			"INSERT OR IGNORE INTO group_members (group_id, user_id) VALUES (?, ?)",
			group.GroupID, member.UserID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func GetUserGroups(userID string) ([]*entity.Group, error) {
	rows, err := DB.Query(`
		SELECT g.group_id, g.group_name, g.date_created 
		FROM groups g
		JOIN group_members gm ON g.group_id = gm.group_id
		WHERE gm.user_id = ?
		ORDER BY g.date_created DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groups := make([]*entity.Group, 0)
	for rows.Next() {
		group := &entity.Group{}
		if err := rows.Scan(&group.GroupID, &group.GroupName, &group.DateCreated); err != nil {
			return nil, err
		}
		group.GroupMembers, _ = GetGroupMembers(group.GroupID)
		groups = append(groups, group)
	}
	return groups, nil
}

func GetUserGroupsWithBalances(userID string) ([]GroupWithBalance, error) {
	rows, err := DB.Query(`
		SELECT g.group_id, g.group_name, g.date_created 
		FROM groups g
		JOIN group_members gm ON g.group_id = gm.group_id
		WHERE gm.user_id = ?
		ORDER BY g.date_created DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groups := make([]GroupWithBalance, 0)
	for rows.Next() {
		group := &entity.Group{}
		if err := rows.Scan(&group.GroupID, &group.GroupName, &group.DateCreated); err != nil {
			return nil, err
		}
		group.GroupMembers, _ = GetGroupMembers(group.GroupID)
		
		// Get balance for this user in this group
		youOwe, youAreOwed := GetUserGroupBalance(userID, group.GroupID)
		
		groups = append(groups, GroupWithBalance{
			Group:      group,
			YouOwe:     youOwe,
			YouAreOwed: youAreOwed,
			HasPending: youOwe > 0 || youAreOwed > 0,
		})
	}
	return groups, nil
}

func GetUserGroupBalance(userID, groupID string) (youOwe float64, youAreOwed float64) {
	// What you owe in this group
	DB.QueryRow(`
		SELECT COALESCE(SUM(amount), 0) 
		FROM balances 
		WHERE group_id = ? AND from_user_id = ? AND amount > 0.10
	`, groupID, userID).Scan(&youOwe)

	// What you're owed in this group
	DB.QueryRow(`
		SELECT COALESCE(SUM(amount), 0) 
		FROM balances 
		WHERE group_id = ? AND to_user_id = ? AND amount > 0.10
	`, groupID, userID).Scan(&youAreOwed)

	return
}

func GetAllGroups() ([]*entity.Group, error) {
	rows, err := DB.Query("SELECT group_id, group_name, date_created FROM groups")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groups := make([]*entity.Group, 0)
	for rows.Next() {
		group := &entity.Group{}
		if err := rows.Scan(&group.GroupID, &group.GroupName, &group.DateCreated); err != nil {
			return nil, err
		}
		group.GroupMembers, _ = GetGroupMembers(group.GroupID)
		groups = append(groups, group)
	}
	return groups, nil
}

func GetGroupByID(groupID string) (*entity.Group, error) {
	group := &entity.Group{}
	err := DB.QueryRow(
		"SELECT group_id, group_name, date_created FROM groups WHERE group_id = ?",
		groupID,
	).Scan(&group.GroupID, &group.GroupName, &group.DateCreated)
	if err != nil {
		return nil, err
	}
	group.GroupMembers, _ = GetGroupMembers(groupID)
	return group, nil
}

func GetGroupMembers(groupID string) ([]*entity.User, error) {
	rows, err := DB.Query(`
		SELECT u.user_id, u.user_name, u.user_email 
		FROM users u 
		JOIN group_members gm ON u.user_id = gm.user_id 
		WHERE gm.group_id = ?
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := make([]*entity.User, 0)
	for rows.Next() {
		user := &entity.User{}
		if err := rows.Scan(&user.UserID, &user.UserName, &user.UserEmail); err != nil {
			return nil, err
		}
		members = append(members, user)
	}
	return members, nil
}

func IsUserInGroup(userID, groupID string) bool {
	var count int
	DB.QueryRow(
		"SELECT COUNT(*) FROM group_members WHERE user_id = ? AND group_id = ?",
		userID, groupID,
	).Scan(&count)
	return count > 0
}

func AddMemberToGroup(groupID, userID string) error {
	_, err := DB.Exec(
		"INSERT OR IGNORE INTO group_members (group_id, user_id) VALUES (?, ?)",
		groupID, userID,
	)
	return err
}
