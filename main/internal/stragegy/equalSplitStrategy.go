package stragegy

import "splitwise/main/internal/entity"

type EqualSplitStrategy struct {
	Group *entity.Group `json:"group"`
}

func NewEqualSplitStrategy(group *entity.Group) *EqualSplitStrategy {
	return &EqualSplitStrategy{
		Group: group,
	}
}

func (e *EqualSplitStrategy) CalculateSplits(splitData map[entity.User]float64, totalAmount float64) []*entity.Split {
	splits := make([]*entity.Split, 0)
	groupMembers := e.Group.GetGroupMembers()
	amountPerPerson := totalAmount / float64(len(groupMembers))
	for _, member := range groupMembers {
		splits = append(splits, entity.NewSplit(member, amountPerPerson))
	}
	return splits
}

func (e *EqualSplitStrategy) GetGroup() *entity.Group {
	return e.Group
}
