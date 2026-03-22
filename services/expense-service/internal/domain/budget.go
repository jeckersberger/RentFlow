package domain

import (
	"fmt"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/events"
)

type BudgetPeriod string

const (
	PeriodMonthly   BudgetPeriod = "monthly"
	PeriodQuarterly BudgetPeriod = "quarterly"
	PeriodAnnual    BudgetPeriod = "annual"
)

type Budget struct {
	events.AggregateRoot
	TenantID   string
	CategoryID string
	ProjectID  string
	Period     BudgetPeriod
	Amount     float64
	Spent      float64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func NewBudget(id, tenantID, categoryID, projectID string, period BudgetPeriod, amount float64) *Budget {
	now := time.Now()
	return &Budget{
		AggregateRoot: *events.NewAggregateRoot(id, "budget"),
		TenantID:      tenantID,
		CategoryID:    categoryID,
		ProjectID:     projectID,
		Period:        period,
		Amount:        amount,
		Spent:         0,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func (b *Budget) AddExpense(amount float64) error {
	if amount < 0 {
		return fmt.Errorf("amount cannot be negative")
	}
	b.Spent += amount
	b.UpdatedAt = time.Now()
	return nil
}

func (b *Budget) IsOverBudget() bool {
	return b.Spent > b.Amount
}

func (b *Budget) RemainingBudget() float64 {
	remaining := b.Amount - b.Spent
	if remaining < 0 {
		return 0
	}
	return remaining
}

func (b *Budget) Validate() error {
	if b.ID == "" {
		return fmt.Errorf("budget ID cannot be empty")
	}
	if b.TenantID == "" {
		return fmt.Errorf("tenant ID cannot be empty")
	}
	if b.Amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	return nil
}
