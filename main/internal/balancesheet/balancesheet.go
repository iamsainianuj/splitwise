package balancesheet

import (
	"fmt"
	"splitwise/main/internal/entity"
)

type BalanceSheet struct {
	Balances map[*entity.User]map[*entity.User]float64
}

func NewBalanceSheet() *BalanceSheet {
	return &BalanceSheet{
		Balances: make(map[*entity.User]map[*entity.User]float64),
	}
}

func (b *BalanceSheet) UpdateBalance(paidBy *entity.User, splits []*entity.Split) {
	for _, split := range splits {
		if b.Balances[paidBy] == nil {
			b.Balances[paidBy] = make(map[*entity.User]float64)
		}
		if b.Balances[split.User] == nil {
			b.Balances[split.User] = make(map[*entity.User]float64)
		}
		b.Balances[paidBy][split.User] -= split.Amount
		b.Balances[split.User][paidBy] += split.Amount
	}
}

func (b *BalanceSheet) PrintBalanceForUser(user *entity.User) {
	for otherUser, amount := range b.Balances[user] {
		if user.UserName != otherUser.UserName {
			fmt.Printf("%s owes %s: %f\n", user.UserName, otherUser.UserName, amount)
		}
	}
}

func (b *BalanceSheet) PrintBalances() {
	for user, balances := range b.Balances {
		fmt.Printf("%s's BalancesSheet:\n", user.UserName)
		for otherUser, amount := range balances {
			fmt.Printf("\t%s: %f\n", otherUser.UserName, amount)
		}
	}
}

func (b *BalanceSheet) SettleBalance(payer *entity.User, payee *entity.User, amount float64) {
	b.Balances[payer][payee] -= amount
	b.Balances[payee][payer] += amount
	if b.Balances[payer][payee] == 0 {
		delete(b.Balances[payer], payee)
		delete(b.Balances[payee], payer)
	}
	if b.Balances[payee][payer] == 0 {
		delete(b.Balances[payee], payer)
	}
	if len(b.Balances[payer]) == 0 {
		delete(b.Balances, payer)
	}
	if len(b.Balances[payee]) == 0 {
		delete(b.Balances, payee)
	}
}
