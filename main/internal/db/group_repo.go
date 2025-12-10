package db

import "splitwise/main/internal/entity"

func CreateGroup(group *entity.Group) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert group
	_, err = tx.Exec(
		"INSERT INTO groups (group_id, group_name) VALUES (?, ?)",
		group.GroupID, group.GroupName,
	)
	if err != nil {
		return err
	}

	// Insert group members
	for _, member := range group.GroupMembers {
		_, err = tx.Exec(
			"INSERT INTO group_members (group_id, user_id) VALUES (?, ?)",
			group.GroupID, member.UserID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
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
		// Load members
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

