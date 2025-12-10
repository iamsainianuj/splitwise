package stragegy

import "splitwise/main/internal/entity"

type ExactSplitStrategy struct {
	Group *entity.Group `json:"group"`
}

func NewExactSplitStrategy(group *entity.Group) *ExactSplitStrategy {
	return &ExactSplitStrategy{
		Group: group,
	}
}
func (e *ExactSplitStrategy) CalculateSplits(splitData map[entity.User]float64, totalAmount float64) []*entity.Split {
	splits := make([]*entity.Split, 0)
	groupMembers := e.Group.GetGroupMembers()
	for _, member := range groupMembers {
		amount := splitData[*member]
		splits = append(splits, entity.NewSplit(member, amount))
	}
	return splits
}
func (e *ExactSplitStrategy) GetGroup() *entity.Group {
	return e.Group
}
