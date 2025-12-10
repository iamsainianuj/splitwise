package stragegy

import "splitwise/main/internal/entity"

type SplitStrategy interface {
	CalculateSplits(splitData map[entity.User]float64, totalAmount float64) []*entity.Split
}
