package service

import (
	"context"
	"errors"
	"time"
)

const (
	WalletKindRecharge       = "recharge"
	WalletKindBonus          = "bonus"
	defaultBonusValidityDays = 90
)

var ErrWalletHoldNotFound = errors.New("wallet hold not found")
var ErrWalletHoldState = errors.New("wallet hold state does not allow this operation")
var ErrWalletBonusGrantNotFound = errors.New("wallet bonus grant not found")
var ErrWalletRefundBonusFrozen = errors.New("refund source bonus is frozen by another operation")

type WalletSummary struct {
	Balance            float64    `json:"balance"`
	AvailableBalance   float64    `json:"available_balance"`
	RechargeBalance    float64    `json:"recharge_balance"`
	BonusBalance       float64    `json:"bonus_balance"`
	OverdraftAmount    float64    `json:"overdraft_amount"`
	FrozenBalance      float64    `json:"frozen_balance"`
	FrozenRecharge     float64    `json:"frozen_recharge_balance"`
	FrozenBonus        float64    `json:"frozen_bonus_balance"`
	TotalBalance       float64    `json:"total_balance"`
	NextBonusExpiresAt *time.Time `json:"next_bonus_expires_at,omitempty"`
	NextExpiringBonus  float64    `json:"next_expiring_bonus_amount"`
}

func NewWalletSummary(balance, bonus, frozen, frozenBonus float64) WalletSummary {
	available := balance
	if available < 0 {
		available = 0
	}
	if bonus < 0 {
		bonus = 0
	}
	if bonus > available {
		bonus = available
	}
	overdraft := 0.0
	if balance < 0 {
		overdraft = -balance
	}
	frozenRecharge := frozen - frozenBonus
	if frozenRecharge < 0 {
		frozenRecharge = 0
	}
	return WalletSummary{
		Balance: balance, AvailableBalance: available,
		RechargeBalance: available - bonus, BonusBalance: bonus, OverdraftAmount: overdraft,
		FrozenBalance: frozen, FrozenRecharge: frozenRecharge, FrozenBonus: frozenBonus,
		TotalBalance: available + frozen,
	}
}

type WalletCreditInput struct {
	UserID           string
	Amount           float64
	Kind             string
	ExpiresAt        *time.Time
	SourceType       string
	SourceID         string
	IdempotencyKey   string
	Notes            string
	CountAsRecharged bool
}

type WalletDebitInput struct {
	UserID         string
	Amount         float64
	AllowOverdraft bool
	SourceType     string
	SourceID       string
	IdempotencyKey string
	Notes          string
}

type WalletSetInput struct {
	UserID         string
	RechargeAmount float64
	BonusAmount    float64
	BonusExpiresAt *time.Time
	SourceType     string
	SourceID       string
	IdempotencyKey string
	Notes          string
}

type WalletMutationResult struct {
	Applied        bool          `json:"applied"`
	Amount         float64       `json:"amount"`
	BonusAmount    float64       `json:"bonus_amount"`
	RechargeAmount float64       `json:"recharge_amount"`
	BonusGrantID   *string       `json:"bonus_grant_id,omitempty"`
	Summary        WalletSummary `json:"summary"`
}

type WalletHoldInput struct {
	UserID             string
	Amount             float64
	Purpose            string
	ReferenceID        string
	RequestFingerprint string
	Notes              string
}

type WalletHoldResult struct {
	Applied        bool          `json:"applied"`
	HoldID         string        `json:"hold_id"`
	Status         string        `json:"status"`
	Amount         float64       `json:"amount"`
	BonusAmount    float64       `json:"bonus_amount"`
	RechargeAmount float64       `json:"recharge_amount"`
	Summary        WalletSummary `json:"summary"`
}

// RefundPointCapacity is the wallet value that can safely be recovered for a
// recharge refund without consuming bonus grants from unrelated sources.
type RefundPointCapacity struct {
	RechargeAvailable    float64 `json:"recharge_available"`
	SourceBonusAvailable float64 `json:"source_bonus_available"`
	SourceBonusExpired   float64 `json:"source_bonus_expired"`
	SourceBonusFrozen    float64 `json:"source_bonus_frozen"`
}

type RefundPointHoldInput struct {
	UserID             string
	RefundID           string
	BonusGrantID       string
	BasePoints         float64
	BonusPoints        float64
	BonusExpiredOffset float64
	RequestFingerprint string
	Notes              string
}

// RefundWalletRepository is deliberately separate from WalletRepository so
// existing consumers and test doubles do not need refund-only methods.
type RefundWalletRepository interface {
	GetRefundPointCapacity(ctx context.Context, userID, bonusGrantID string) (RefundPointCapacity, error)
	HoldRefundPoints(ctx context.Context, input RefundPointHoldInput) (WalletHoldResult, error)
}

type WalletTransaction struct {
	ID             string    `json:"id"`
	Action         string    `json:"action"`
	Amount         float64   `json:"amount"`
	BonusAmount    float64   `json:"bonus_amount"`
	RechargeAmount float64   `json:"recharge_amount"`
	FrozenAmount   float64   `json:"frozen_amount"`
	BalanceBefore  float64   `json:"balance_before"`
	BalanceAfter   float64   `json:"balance_after"`
	BonusBefore    float64   `json:"bonus_before"`
	BonusAfter     float64   `json:"bonus_after"`
	SourceType     string    `json:"source_type"`
	SourceID       string    `json:"source_id"`
	Notes          string    `json:"notes"`
	CreatedAt      time.Time `json:"created_at"`
}

type WalletTransactionPage struct {
	Items []WalletTransaction `json:"items"`
	Total int64               `json:"total"`
}

type WalletRepository interface {
	GetWalletSummary(ctx context.Context, userID string) (WalletSummary, error)
	CreditWallet(ctx context.Context, input WalletCreditInput) (WalletMutationResult, error)
	DebitWallet(ctx context.Context, input WalletDebitInput) (WalletMutationResult, error)
	SetWalletBalance(ctx context.Context, input WalletSetInput) (WalletMutationResult, error)
	HoldWallet(ctx context.Context, input WalletHoldInput) (WalletHoldResult, error)
	CaptureWalletHold(ctx context.Context, holdID, idempotencyKey string) (WalletHoldResult, error)
	ReleaseWalletHold(ctx context.Context, holdID, idempotencyKey string) (WalletHoldResult, error)
	RefundWalletHold(ctx context.Context, holdID string, amount float64, idempotencyKey string) (WalletMutationResult, error)
	ListWalletTransactions(ctx context.Context, userID string, page, pageSize int) (WalletTransactionPage, error)
	ExpireBonusBalances(ctx context.Context, limit int) ([]string, error)
}
