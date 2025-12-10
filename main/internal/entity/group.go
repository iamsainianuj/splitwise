package entity

import "time"

type Group struct {
	GroupID      string    `json:"group_id"`
	GroupName    string    `json:"group_name"`
	GroupMembers []*User   `json:"group_members"`
	DateCreated  time.Time `json:"date_created"`
}

func NewGroup(groupID, groupName string, groupMembers []*User) *Group {
	return &Group{
		GroupID:      groupID,
		GroupName:    groupName,
		GroupMembers: groupMembers,
	}
}

func (g *Group) GetGroupID() string {
	return g.GroupID
}
func (g *Group) GetGroupName() string {
	return g.GroupName
}
func (g *Group) GetGroupMembers() []*User {
	return g.GroupMembers
}
func (g *Group) GetDateCreated() time.Time {
	return g.DateCreated
}
