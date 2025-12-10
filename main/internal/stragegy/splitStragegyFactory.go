package stragegy

import "splitwise/main/internal/entity"

func GetSplitStrategy(splitType SplitType, group *entity.Group) SplitStrategy {
	switch splitType {
	case Equal:
		return NewEqualSplitStrategy(group)
	case Percentage:
		return NewPercentageSplitStrategy(group)
	case Exact:
		return NewExactSplitStrategy(group)
	}
	return nil
}
