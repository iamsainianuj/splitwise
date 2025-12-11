package stragegy

import (
	"math"
	"splitwise/main/internal/entity"
)

type PercentageSplitStrategy struct {
	Group *entity.Group `json:"group"`
}

func NewPercentageSplitStrategy(group *entity.Group) *PercentageSplitStrategy {
	return &PercentageSplitStrategy{
		Group: group,
	}
}
func (p *PercentageSplitStrategy) CalculateSplits(splitData map[entity.User]float64, totalAmount float64) []*entity.Split {
	splits := make([]*entity.Split, 0)
	groupMembers := p.Group.GetGroupMembers()
	myPart := 100 / len(groupMembers)
	for _, member := range groupMembers {
		percentage := math.Abs(splitData[*member] - float64(myPart))
		amount := totalAmount * percentage / 100
		splits = append(splits, entity.NewSplit(member, amount))
	}
	return splits
}
func (p *PercentageSplitStrategy) GetGroup() *entity.Group {
	return p.Group
}
