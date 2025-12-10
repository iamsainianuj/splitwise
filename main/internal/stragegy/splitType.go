package stragegy

type SplitType int

const (
	Equal SplitType = iota
	Percentage
	Exact
)

func (s SplitType) String() string {
	return []string{"Equal", "Percentage", "Exact"}[s]
}
func (s SplitType) GetSplitType() SplitType {
	return []SplitType{Equal, Percentage, Exact}[s]
}
