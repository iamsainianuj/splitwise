package service

import (
	"splitwise/main/internal/balancesheet"
	"splitwise/main/internal/entity"
	"splitwise/main/internal/expense"
	"splitwise/main/internal/stragegy"
)

type SplitWiseService struct {
	Users        []*entity.User
	Groups       []*entity.Group
	Expenses     []*expense.Expense
	BalanceSheet *balancesheet.BalanceSheet
}

func NewSplitWiseService() *SplitWiseService {
	return &SplitWiseService{
		Users:        make([]*entity.User, 0),
		Groups:       make([]*entity.Group, 0),
		Expenses:     make([]*expense.Expense, 0),
		BalanceSheet: balancesheet.NewBalanceSheet(),
	}
}

func (s *SplitWiseService) CreateUser(userID, userName, userEmail string) *entity.User {
	user := entity.NewUser(userID, userName, userEmail)
	s.Users = append(s.Users, user)
	return user
}

func (s *SplitWiseService) CreateGroup(groupID, groupName string, groupMembers []*entity.User) *entity.Group {
	group := entity.NewGroup(groupID, groupName, groupMembers)
	s.Groups = append(s.Groups, group)
	return group
}

func (s *SplitWiseService) AddExpense(expenseID, expenseDescription string, expenseAmount float64, paidByUserID string, groupID string, splitType stragegy.SplitType, splitData map[entity.User]float64) *expense.Expense {
	paidBy := s.getUserByID(paidByUserID)
	group := s.getGroupByID(groupID)
	splits := stragegy.GetSplitStrategy(splitType, group).CalculateSplits(splitData, expenseAmount)
	if paidBy == nil || group == nil || splits == nil {
		return nil
	}
	expense := expense.NewExpense(expenseID, expenseDescription, expenseAmount, group, paidBy, splits)
	s.BalanceSheet.UpdateBalance(paidBy, splits)
	s.Expenses = append(s.Expenses, expense)
	return expense
}
func (s *SplitWiseService) getUserByID(userID string) *entity.User {
	for _, user := range s.Users {
		if user.UserID == userID {
			return user
		}
	}
	return nil
}
func (s *SplitWiseService) getGroupByID(groupID string) *entity.Group {
	for _, group := range s.Groups {
		if group.GroupID == groupID {
			return group
		}
	}
	return nil
}

func (s *SplitWiseService) Settle(fromUserID, toUserID string, amount float64) {
	fromUser := s.getUserByID(fromUserID)
	toUser := s.getUserByID(toUserID)
	if fromUser == nil || toUser == nil {
		return
	}
	s.BalanceSheet.SettleBalance(fromUser, toUser, amount)
}

func (s *SplitWiseService) PrintBalances() {
	s.BalanceSheet.PrintBalances()
}

func (s *SplitWiseService) PrintBalanceForUser(userID string) {
	user := s.getUserByID(userID)
	if user == nil {
		return
	}
	s.BalanceSheet.PrintBalanceForUser(user)
}

func (s *SplitWiseService) DeleteExpense(expenseID string) {
	for i, expense := range s.Expenses {
		if expense.GetExpenseID() == expenseID {
			s.Expenses = append(s.Expenses[:i], s.Expenses[i+1:]...)
			break
		}
	}
}
