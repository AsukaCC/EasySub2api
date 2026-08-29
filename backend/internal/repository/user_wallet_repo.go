package repository

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	dbent "github.com/AsukaCC/EasySub2api/ent"
	"github.com/AsukaCC/EasySub2api/internal/service"
	"github.com/google/uuid"
)

const defaultBonusValidityDays = 90

type walletUserRow struct {
	balance     float64
	bonus       float64
	frozen      float64
	frozenBonus float64
}

func walletMoney(value float64) float64 {
	return service.QuantizeUsageBillingAmount(value)
}

func newWalletUUID() string {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.NewString()
	}
	return id.String()
}

func walletIdempotencyKey(prefix, provided string) string {
	if value := strings.TrimSpace(provided); value != "" {
		return value
	}
	return prefix + ":" + newWalletUUID()
}

func (r *userRepository) withWalletTx(ctx context.Context, fn func(context.Context, sqlQueryExecutor) error) error {
	if dbent.TxFromContext(ctx) != nil {
		exec := txAwareSQLExecutor(ctx, r.sql, r.client)
		if exec == nil {
			return errors.New("wallet SQL executor is not configured")
		}
		return fn(ctx, exec)
	}
	tx, err := r.client.Tx(ctx)
	if errors.Is(err, dbent.ErrTxStarted) {
		exec := txAwareSQLExecutor(ctx, r.sql, r.client)
		if exec == nil {
			return errors.New("wallet SQL executor is not configured")
		}
		return fn(ctx, exec)
	}
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)
	exec := txAwareSQLExecutor(txCtx, r.sql, r.client)
	if exec == nil {
		return errors.New("wallet transaction SQL executor is not configured")
	}
	if err := fn(txCtx, exec); err != nil {
		return err
	}
	return tx.Commit()
}

func lockWalletUser(ctx context.Context, exec sqlQueryExecutor, userID string) (walletUserRow, error) {
	rows, err := exec.QueryContext(ctx, `
		SELECT balance, bonus_balance, frozen_balance, frozen_bonus_balance
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
		FOR UPDATE
	`, userID)
	if err != nil {
		return walletUserRow{}, err
	}
	defer rows.Close() //nolint:errcheck
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return walletUserRow{}, err
		}
		return walletUserRow{}, service.ErrUserNotFound
	}
	var row walletUserRow
	if err := rows.Scan(&row.balance, &row.bonus, &row.frozen, &row.frozenBonus); err != nil {
		return walletUserRow{}, err
	}
	return row, rows.Err()
}

func loadWalletSummary(ctx context.Context, exec sqlQueryExecutor, userID string) (service.WalletSummary, error) {
	rows, err := exec.QueryContext(ctx, `
		SELECT balance, bonus_balance, frozen_balance, frozen_bonus_balance
		FROM users WHERE id = $1 AND deleted_at IS NULL
	`, userID)
	if err != nil {
		return service.WalletSummary{}, err
	}
	var balance, bonus, frozen, frozenBonus float64
	if !rows.Next() {
		_ = rows.Close()
		return service.WalletSummary{}, service.ErrUserNotFound
	}
	if err := rows.Scan(&balance, &bonus, &frozen, &frozenBonus); err != nil {
		_ = rows.Close()
		return service.WalletSummary{}, err
	}
	if err := rows.Close(); err != nil {
		return service.WalletSummary{}, err
	}
	summary := service.NewWalletSummary(balance, bonus, frozen, frozenBonus)
	expiryRows, err := exec.QueryContext(ctx, `
		SELECT expires_at, SUM(remaining_amount)
		FROM wallet_bonus_grants
		WHERE user_id = $1 AND remaining_amount > 0 AND expires_at > NOW()
		GROUP BY expires_at
		ORDER BY expires_at ASC
		LIMIT 1
	`, userID)
	if err != nil {
		return service.WalletSummary{}, err
	}
	defer expiryRows.Close() //nolint:errcheck
	if expiryRows.Next() {
		var expiresAt time.Time
		if err := expiryRows.Scan(&expiresAt, &summary.NextExpiringBonus); err != nil {
			return service.WalletSummary{}, err
		}
		summary.NextBonusExpiresAt = &expiresAt
	}
	return summary, expiryRows.Err()
}

func walletTransactionExists(ctx context.Context, exec sqlQueryExecutor, key string) (bool, error) {
	rows, err := exec.QueryContext(ctx, `SELECT 1 FROM wallet_transactions WHERE idempotency_key = $1`, key)
	if err != nil {
		return false, err
	}
	defer rows.Close() //nolint:errcheck
	return rows.Next(), rows.Err()
}

func insertWalletTransaction(ctx context.Context, exec sqlQueryExecutor, userID, action string, amount, bonusAmount, rechargeAmount, frozenAmount float64, before, after walletUserRow, sourceType, sourceID, key, notes string) error {
	_, err := exec.ExecContext(ctx, `
		INSERT INTO wallet_transactions (
			user_id, action, amount, bonus_amount, recharge_amount, frozen_amount,
			balance_before, balance_after, bonus_before, bonus_after,
			source_type, source_id, idempotency_key, notes
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
	`, userID, action, walletMoney(amount), walletMoney(bonusAmount), walletMoney(rechargeAmount), walletMoney(frozenAmount),
		walletMoney(before.balance), walletMoney(after.balance), walletMoney(before.bonus), walletMoney(after.bonus),
		strings.TrimSpace(sourceType), strings.TrimSpace(sourceID), key, strings.TrimSpace(notes))
	return err
}

func expireUserBonusTx(ctx context.Context, exec sqlQueryExecutor, userID string, row *walletUserRow) (float64, error) {
	rows, err := exec.QueryContext(ctx, `
		SELECT id, remaining_amount
		FROM wallet_bonus_grants
		WHERE user_id = $1 AND remaining_amount > 0 AND expires_at <= NOW()
		ORDER BY expires_at ASC, created_at ASC, id ASC
		FOR UPDATE
	`, userID)
	if err != nil {
		return 0, err
	}
	type expiredGrant struct {
		id     string
		amount float64
	}
	grants := make([]expiredGrant, 0)
	var total float64
	for rows.Next() {
		var grant expiredGrant
		if err := rows.Scan(&grant.id, &grant.amount); err != nil {
			_ = rows.Close()
			return 0, err
		}
		grant.amount = walletMoney(grant.amount)
		total = walletMoney(total + grant.amount)
		grants = append(grants, grant)
	}
	if err := rows.Close(); err != nil {
		return 0, err
	}
	if total <= 0 {
		return 0, nil
	}
	before := *row
	for _, grant := range grants {
		if _, err := exec.ExecContext(ctx, `
			UPDATE wallet_bonus_grants
			SET remaining_amount = 0,
				expired_amount = expired_amount + $2,
				status = CASE WHEN frozen_amount > 0 THEN 'held_expired' ELSE 'expired' END,
				updated_at = NOW()
			WHERE id = $1
		`, grant.id, grant.amount); err != nil {
			return 0, err
		}
	}
	if _, err := exec.ExecContext(ctx, `
		UPDATE users
		SET balance = balance - $1, bonus_balance = bonus_balance - $1, updated_at = NOW()
		WHERE id = $2
	`, total, userID); err != nil {
		return 0, err
	}
	row.balance = walletMoney(row.balance - total)
	row.bonus = walletMoney(row.bonus - total)
	return total, insertWalletTransaction(ctx, exec, userID, "expire", -total, -total, 0, 0, before, *row, "bonus_expiry", "", "bonus-expiry:"+newWalletUUID(), "expired bonus balance")
}

func consumeBonusTx(ctx context.Context, exec sqlQueryExecutor, userID string, amount float64, freeze bool) (float64, map[string]float64, error) {
	rows, err := exec.QueryContext(ctx, `
		SELECT id, remaining_amount
		FROM wallet_bonus_grants
		WHERE user_id = $1 AND remaining_amount > 0 AND expires_at > NOW()
		ORDER BY expires_at ASC, created_at ASC, id ASC
		FOR UPDATE
	`, userID)
	if err != nil {
		return 0, nil, err
	}
	type bonusAllocation struct {
		grantID string
		amount  float64
	}
	allocations := make(map[string]float64)
	planned := make([]bonusAllocation, 0)
	remaining := walletMoney(amount)
	var used float64
	for rows.Next() && remaining > 0 {
		var id string
		var available float64
		if err := rows.Scan(&id, &available); err != nil {
			_ = rows.Close()
			return 0, nil, err
		}
		use := math.Min(walletMoney(available), remaining)
		use = walletMoney(use)
		if use <= 0 {
			continue
		}
		planned = append(planned, bonusAllocation{grantID: id, amount: use})
		allocations[id] = use
		used = walletMoney(used + use)
		remaining = walletMoney(remaining - use)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return 0, nil, err
	}
	if err := rows.Close(); err != nil {
		return 0, nil, err
	}

	for _, allocation := range planned {
		if freeze {
			_, err = exec.ExecContext(ctx, `
				UPDATE wallet_bonus_grants
				SET remaining_amount = remaining_amount - $1,
					frozen_amount = frozen_amount + $1,
					status = 'held', updated_at = NOW()
				WHERE id = $2
			`, allocation.amount, allocation.grantID)
		} else {
			_, err = exec.ExecContext(ctx, `
				UPDATE wallet_bonus_grants
				SET remaining_amount = remaining_amount - $1,
					spent_amount = spent_amount + $1,
					status = CASE WHEN remaining_amount - $1 = 0 AND frozen_amount = 0 THEN 'spent' ELSE status END,
					updated_at = NOW()
				WHERE id = $2
			`, allocation.amount, allocation.grantID)
		}
		if err != nil {
			return 0, nil, err
		}
	}
	return used, allocations, nil
}

func (r *userRepository) GetWalletSummary(ctx context.Context, userID string) (service.WalletSummary, error) {
	var summary service.WalletSummary
	err := r.withWalletTx(ctx, func(txCtx context.Context, exec sqlQueryExecutor) error {
		row, err := lockWalletUser(txCtx, exec, userID)
		if err != nil {
			return err
		}
		if _, err := expireUserBonusTx(txCtx, exec, userID, &row); err != nil {
			return err
		}
		summary, err = loadWalletSummary(txCtx, exec, userID)
		return err
	})
	return summary, err
}

func (r *userRepository) CreditWallet(ctx context.Context, input service.WalletCreditInput) (service.WalletMutationResult, error) {
	amount := walletMoney(input.Amount)
	if amount <= 0 || math.IsNaN(amount) || math.IsInf(amount, 0) {
		return service.WalletMutationResult{}, fmt.Errorf("wallet credit amount must be positive")
	}
	kind := strings.TrimSpace(input.Kind)
	if kind == "" {
		kind = service.WalletKindRecharge
	}
	if kind != service.WalletKindRecharge && kind != service.WalletKindBonus {
		return service.WalletMutationResult{}, fmt.Errorf("unsupported wallet credit kind %q", kind)
	}
	expiresAt := input.ExpiresAt
	if kind == service.WalletKindBonus {
		if expiresAt == nil {
			value := time.Now().UTC().Add(defaultBonusValidityDays * 24 * time.Hour)
			expiresAt = &value
		}
		if !expiresAt.After(time.Now().UTC()) {
			return service.WalletMutationResult{}, fmt.Errorf("bonus balance expiration must be in the future")
		}
	}
	key := walletIdempotencyKey("wallet-credit", input.IdempotencyKey)
	result := service.WalletMutationResult{Amount: amount}
	err := r.withWalletTx(ctx, func(txCtx context.Context, exec sqlQueryExecutor) error {
		exists, err := walletTransactionExists(txCtx, exec, key)
		if err != nil {
			return err
		}
		if exists {
			if kind == service.WalletKindBonus {
				result.BonusGrantID, err = findWalletBonusGrantID(txCtx, exec, input.UserID, input.SourceType, input.SourceID)
				if err != nil {
					return err
				}
			}
			result.Summary, err = loadWalletSummary(txCtx, exec, input.UserID)
			return err
		}
		row, err := lockWalletUser(txCtx, exec, input.UserID)
		if err != nil {
			return err
		}
		if _, err := expireUserBonusTx(txCtx, exec, input.UserID, &row); err != nil {
			return err
		}
		before := row
		bonusAvailable := 0.0
		if kind == service.WalletKindBonus {
			bonusAvailable = walletMoney(math.Min(amount, math.Max(row.balance+amount, 0)))
		}
		if _, err := exec.ExecContext(txCtx, `
			UPDATE users
			SET balance = balance + $1,
				bonus_balance = bonus_balance + $2,
				total_recharged = total_recharged + $3,
				updated_at = NOW()
			WHERE id = $4
		`, amount, bonusAvailable, func() float64 {
			if input.CountAsRecharged {
				return amount
			}
			return 0
		}(), input.UserID); err != nil {
			return err
		}
		row.balance = walletMoney(row.balance + amount)
		row.bonus = walletMoney(row.bonus + bonusAvailable)
		if kind == service.WalletKindBonus {
			status := "active"
			if bonusAvailable <= 0 {
				status = "spent"
			}
			spentAmount := walletMoney(amount - bonusAvailable)
			grantID := newWalletUUID()
			if _, err := exec.ExecContext(txCtx, `
				INSERT INTO wallet_bonus_grants (
					id, user_id, original_amount, remaining_amount, spent_amount, expires_at, source_type, source_id, status
				) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
			`, grantID, input.UserID, amount, bonusAvailable, spentAmount, expiresAt.UTC(), input.SourceType, input.SourceID, status); err != nil {
				return err
			}
			result.BonusGrantID = &grantID
			result.BonusAmount = bonusAvailable
		} else {
			result.RechargeAmount = amount
		}
		if err := insertWalletTransaction(txCtx, exec, input.UserID, "credit", amount, result.BonusAmount, result.RechargeAmount, 0, before, row, input.SourceType, input.SourceID, key, input.Notes); err != nil {
			return err
		}
		result.Applied = true
		result.Summary, err = loadWalletSummary(txCtx, exec, input.UserID)
		return err
	})
	return result, err
}

func findWalletBonusGrantID(ctx context.Context, exec sqlQueryExecutor, userID, sourceType, sourceID string) (*string, error) {
	rows, err := exec.QueryContext(ctx, `
		SELECT id FROM wallet_bonus_grants
		WHERE user_id = $1 AND source_type = $2 AND source_id = $3
		LIMIT 1
	`, userID, sourceType, sourceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck
	if !rows.Next() {
		return nil, rows.Err()
	}
	var grantID string
	if err := rows.Scan(&grantID); err != nil {
		return nil, err
	}
	return &grantID, rows.Err()
}

func (r *userRepository) DebitWallet(ctx context.Context, input service.WalletDebitInput) (service.WalletMutationResult, error) {
	amount := walletMoney(input.Amount)
	if amount <= 0 || math.IsNaN(amount) || math.IsInf(amount, 0) {
		return service.WalletMutationResult{}, fmt.Errorf("wallet debit amount must be positive")
	}
	key := walletIdempotencyKey("wallet-debit", input.IdempotencyKey)
	result := service.WalletMutationResult{Amount: amount}
	err := r.withWalletTx(ctx, func(txCtx context.Context, exec sqlQueryExecutor) error {
		var err error
		result, err = debitWalletTx(txCtx, exec, input, key)
		return err
	})
	return result, err
}

func (r *userRepository) SetWalletBalance(ctx context.Context, input service.WalletSetInput) (service.WalletMutationResult, error) {
	recharge := walletMoney(input.RechargeAmount)
	bonus := walletMoney(input.BonusAmount)
	if recharge < 0 || bonus < 0 {
		return service.WalletMutationResult{}, fmt.Errorf("wallet bucket amounts cannot be negative")
	}
	if bonus > 0 && (input.BonusExpiresAt == nil || !input.BonusExpiresAt.After(time.Now().UTC())) {
		return service.WalletMutationResult{}, fmt.Errorf("bonus balance expiration must be in the future")
	}
	key := walletIdempotencyKey("wallet-set", input.IdempotencyKey)
	result := service.WalletMutationResult{Amount: walletMoney(recharge + bonus), RechargeAmount: recharge, BonusAmount: bonus}
	err := r.withWalletTx(ctx, func(txCtx context.Context, exec sqlQueryExecutor) error {
		exists, err := walletTransactionExists(txCtx, exec, key)
		if err != nil {
			return err
		}
		if exists {
			result.Summary, err = loadWalletSummary(txCtx, exec, input.UserID)
			return err
		}
		row, err := lockWalletUser(txCtx, exec, input.UserID)
		if err != nil {
			return err
		}
		if _, err := expireUserBonusTx(txCtx, exec, input.UserID, &row); err != nil {
			return err
		}
		before := row
		if _, err := exec.ExecContext(txCtx, `
			UPDATE wallet_bonus_grants SET reversed_amount = reversed_amount + remaining_amount,
				remaining_amount = 0,
				status = CASE WHEN frozen_amount > 0 THEN 'held_adjusted' ELSE 'adjusted' END, updated_at = NOW()
			WHERE user_id = $1 AND remaining_amount > 0
		`, input.UserID); err != nil {
			return err
		}
		if bonus > 0 {
			grantID := newWalletUUID()
			if _, err := exec.ExecContext(txCtx, `
				INSERT INTO wallet_bonus_grants
					(id, user_id, original_amount, remaining_amount, frozen_amount, source_type, source_id, status, expires_at, created_at, updated_at)
				VALUES ($1, $2, $3, $3, 0, $4, $5, 'active', $6, NOW(), NOW())
			`, grantID, input.UserID, bonus, input.SourceType, input.SourceID, *input.BonusExpiresAt); err != nil {
				return err
			}
		}
		if _, err := exec.ExecContext(txCtx, `
			UPDATE users SET balance = $1, bonus_balance = $2, updated_at = NOW() WHERE id = $3
		`, recharge+bonus, bonus, input.UserID); err != nil {
			return err
		}
		row.balance = recharge + bonus
		row.bonus = bonus
		if err := insertWalletTransaction(txCtx, exec, input.UserID, "adjust", row.balance-before.balance,
			row.bonus-before.bonus, recharge-math.Max(before.balance-before.bonus, 0), 0,
			before, row, input.SourceType, input.SourceID, key, input.Notes); err != nil {
			return err
		}
		result.Applied = true
		result.Summary, err = loadWalletSummary(txCtx, exec, input.UserID)
		return err
	})
	return result, err
}

func debitWalletTx(ctx context.Context, exec sqlQueryExecutor, input service.WalletDebitInput, key string) (service.WalletMutationResult, error) {
	amount := walletMoney(input.Amount)
	result := service.WalletMutationResult{Amount: amount}
	exists, err := walletTransactionExists(ctx, exec, key)
	if err != nil {
		return result, err
	}
	if exists {
		result.Summary, err = loadWalletSummary(ctx, exec, input.UserID)
		return result, err
	}
	row, err := lockWalletUser(ctx, exec, input.UserID)
	if err != nil {
		return result, err
	}
	if _, err := expireUserBonusTx(ctx, exec, input.UserID, &row); err != nil {
		return result, err
	}
	if !input.AllowOverdraft && row.balance < amount {
		return result, service.ErrInsufficientBalance
	}
	before := row
	bonusUsed, _, err := consumeBonusTx(ctx, exec, input.UserID, math.Min(amount, math.Max(row.bonus, 0)), false)
	if err != nil {
		return result, err
	}
	rechargeUsed := walletMoney(amount - bonusUsed)
	if _, err := exec.ExecContext(ctx, `
		UPDATE users SET balance = balance - $1, bonus_balance = bonus_balance - $2, updated_at = NOW()
		WHERE id = $3
	`, amount, bonusUsed, input.UserID); err != nil {
		return result, err
	}
	row.balance = walletMoney(row.balance - amount)
	row.bonus = walletMoney(row.bonus - bonusUsed)
	result.BonusAmount = bonusUsed
	result.RechargeAmount = rechargeUsed
	if err := insertWalletTransaction(ctx, exec, input.UserID, "debit", -amount, -bonusUsed, -rechargeUsed, 0, before, row, input.SourceType, input.SourceID, key, input.Notes); err != nil {
		return result, err
	}
	result.Applied = true
	result.Summary, err = loadWalletSummary(ctx, exec, input.UserID)
	return result, err
}

func (r *userRepository) HoldWallet(ctx context.Context, input service.WalletHoldInput) (service.WalletHoldResult, error) {
	amount := walletMoney(input.Amount)
	if amount <= 0 {
		return service.WalletHoldResult{}, fmt.Errorf("wallet hold amount must be positive")
	}
	result := service.WalletHoldResult{Amount: amount}
	err := r.withWalletTx(ctx, func(txCtx context.Context, exec sqlQueryExecutor) error {
		rows, err := exec.QueryContext(txCtx, `
			SELECT id, status, amount, bonus_amount, recharge_amount, request_fingerprint
			FROM wallet_holds WHERE purpose = $1 AND reference_id = $2 FOR UPDATE
		`, input.Purpose, input.ReferenceID)
		if err != nil {
			return err
		}
		if rows.Next() {
			var fingerprint string
			if err := rows.Scan(&result.HoldID, &result.Status, &result.Amount, &result.BonusAmount, &result.RechargeAmount, &fingerprint); err != nil {
				_ = rows.Close()
				return err
			}
			_ = rows.Close()
			if fingerprint != "" && input.RequestFingerprint != "" && fingerprint != input.RequestFingerprint {
				return service.ErrUsageBillingRequestConflict
			}
			result.Summary, err = loadWalletSummary(txCtx, exec, input.UserID)
			return err
		}
		if err := rows.Close(); err != nil {
			return err
		}
		row, err := lockWalletUser(txCtx, exec, input.UserID)
		if err != nil {
			return err
		}
		if _, err := expireUserBonusTx(txCtx, exec, input.UserID, &row); err != nil {
			return err
		}
		if row.balance < amount {
			return service.ErrInsufficientBalance
		}
		before := row
		bonusUsed, allocations, err := consumeBonusTx(txCtx, exec, input.UserID, math.Min(amount, row.bonus), true)
		if err != nil {
			return err
		}
		rechargeUsed := walletMoney(amount - bonusUsed)
		holdID := newWalletUUID()
		if _, err := exec.ExecContext(txCtx, `
			INSERT INTO wallet_holds (id, user_id, purpose, reference_id, amount, bonus_amount, recharge_amount, request_fingerprint)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		`, holdID, input.UserID, input.Purpose, input.ReferenceID, amount, bonusUsed, rechargeUsed, input.RequestFingerprint); err != nil {
			return err
		}
		for grantID, allocated := range allocations {
			if _, err := exec.ExecContext(txCtx, `
				INSERT INTO wallet_hold_allocations (hold_id, bonus_grant_id, amount)
				VALUES ($1,$2,$3)
			`, holdID, grantID, allocated); err != nil {
				return err
			}
		}
		if _, err := exec.ExecContext(txCtx, `
			UPDATE users SET
				balance = balance - $1, bonus_balance = bonus_balance - $2,
				frozen_balance = frozen_balance + $1, frozen_bonus_balance = frozen_bonus_balance + $2,
				updated_at = NOW()
			WHERE id = $3
		`, amount, bonusUsed, input.UserID); err != nil {
			return err
		}
		row.balance = walletMoney(row.balance - amount)
		row.bonus = walletMoney(row.bonus - bonusUsed)
		row.frozen = walletMoney(row.frozen + amount)
		row.frozenBonus = walletMoney(row.frozenBonus + bonusUsed)
		if err := insertWalletTransaction(txCtx, exec, input.UserID, "hold", -amount, -bonusUsed, -rechargeUsed, amount, before, row, input.Purpose, input.ReferenceID, "wallet-hold:"+input.Purpose+":"+input.ReferenceID, input.Notes); err != nil {
			return err
		}
		result.Applied = true
		result.HoldID = holdID
		result.Status = "held"
		result.BonusAmount = bonusUsed
		result.RechargeAmount = rechargeUsed
		result.Summary, err = loadWalletSummary(txCtx, exec, input.UserID)
		return err
	})
	return result, err
}

func (r *userRepository) GetRefundPointCapacity(ctx context.Context, userID, bonusGrantID string) (service.RefundPointCapacity, error) {
	var capacity service.RefundPointCapacity
	err := r.withWalletTx(ctx, func(txCtx context.Context, exec sqlQueryExecutor) error {
		row, err := lockWalletUser(txCtx, exec, userID)
		if err != nil {
			return err
		}
		if _, err := expireUserBonusTx(txCtx, exec, userID, &row); err != nil {
			return err
		}
		capacity.RechargeAvailable = walletMoney(math.Max(row.balance-row.bonus, 0))
		bonusGrantID = strings.TrimSpace(bonusGrantID)
		if bonusGrantID == "" {
			return nil
		}
		rows, err := exec.QueryContext(txCtx, `
			SELECT remaining_amount, frozen_amount, expired_amount
			FROM wallet_bonus_grants
			WHERE id = $1 AND user_id = $2
			FOR UPDATE
		`, bonusGrantID, userID)
		if err != nil {
			return err
		}
		defer rows.Close() //nolint:errcheck
		if !rows.Next() {
			if err := rows.Err(); err != nil {
				return err
			}
			return service.ErrWalletBonusGrantNotFound
		}
		if err := rows.Scan(&capacity.SourceBonusAvailable, &capacity.SourceBonusFrozen, &capacity.SourceBonusExpired); err != nil {
			return err
		}
		capacity.SourceBonusAvailable = walletMoney(capacity.SourceBonusAvailable)
		capacity.SourceBonusFrozen = walletMoney(capacity.SourceBonusFrozen)
		capacity.SourceBonusExpired = walletMoney(capacity.SourceBonusExpired)
		return rows.Err()
	})
	return capacity, err
}

func (r *userRepository) HoldRefundPoints(ctx context.Context, input service.RefundPointHoldInput) (service.WalletHoldResult, error) {
	basePoints := walletMoney(input.BasePoints)
	bonusPoints := walletMoney(input.BonusPoints)
	expiredOffset := walletMoney(input.BonusExpiredOffset)
	if basePoints < 0 || bonusPoints < 0 || expiredOffset < 0 || expiredOffset > bonusPoints {
		return service.WalletHoldResult{}, fmt.Errorf("invalid refund point hold amounts")
	}
	amount := walletMoney(basePoints + bonusPoints - expiredOffset)
	if amount <= 0 {
		return service.WalletHoldResult{Status: "not_required"}, nil
	}
	if strings.TrimSpace(input.RefundID) == "" {
		return service.WalletHoldResult{}, fmt.Errorf("refund point hold requires refund id")
	}

	result := service.WalletHoldResult{Amount: amount}
	err := r.withWalletTx(ctx, func(txCtx context.Context, exec sqlQueryExecutor) error {
		rows, err := exec.QueryContext(txCtx, `
			SELECT id, status, amount, bonus_amount, recharge_amount, request_fingerprint
			FROM wallet_holds WHERE purpose = 'payment_refund' AND reference_id = $1 FOR UPDATE
		`, input.RefundID)
		if err != nil {
			return err
		}
		if rows.Next() {
			var fingerprint string
			if err := rows.Scan(&result.HoldID, &result.Status, &result.Amount, &result.BonusAmount, &result.RechargeAmount, &fingerprint); err != nil {
				_ = rows.Close()
				return err
			}
			_ = rows.Close()
			if fingerprint != "" && input.RequestFingerprint != "" && fingerprint != input.RequestFingerprint {
				return service.ErrUsageBillingRequestConflict
			}
			result.Summary, err = loadWalletSummary(txCtx, exec, input.UserID)
			return err
		}
		if err := rows.Close(); err != nil {
			return err
		}

		row, err := lockWalletUser(txCtx, exec, input.UserID)
		if err != nil {
			return err
		}
		if _, err := expireUserBonusTx(txCtx, exec, input.UserID, &row); err != nil {
			return err
		}

		bonusToHold := walletMoney(bonusPoints - expiredOffset)
		sourceBonus := 0.0
		bonusGrantID := strings.TrimSpace(input.BonusGrantID)
		if bonusGrantID != "" && bonusToHold > 0 {
			grantRows, err := exec.QueryContext(txCtx, `
				SELECT remaining_amount
				FROM wallet_bonus_grants
				WHERE id = $1 AND user_id = $2
				FOR UPDATE
			`, bonusGrantID, input.UserID)
			if err != nil {
				return err
			}
			if !grantRows.Next() {
				_ = grantRows.Close()
				return service.ErrWalletBonusGrantNotFound
			}
			var remaining float64
			if err := grantRows.Scan(&remaining); err != nil {
				_ = grantRows.Close()
				return err
			}
			if err := grantRows.Close(); err != nil {
				return err
			}
			// remaining_amount excludes points already frozen by other holds. Use
			// the available portion and recover any bonus shortfall from recharge
			// points below instead of blocking the entire refund.
			sourceBonus = walletMoney(math.Min(walletMoney(remaining), bonusToHold))
		}

		rechargePoints := walletMoney(basePoints + bonusToHold - sourceBonus)
		availableRecharge := walletMoney(math.Max(row.balance-row.bonus, 0))
		if rechargePoints > availableRecharge || amount > walletMoney(math.Max(row.balance, 0)) {
			return service.ErrInsufficientBalance
		}

		before := row
		holdID := newWalletUUID()
		if _, err := exec.ExecContext(txCtx, `
			INSERT INTO wallet_holds (id, user_id, purpose, reference_id, amount, bonus_amount, recharge_amount, request_fingerprint)
			VALUES ($1,$2,'payment_refund',$3,$4,$5,$6,$7)
		`, holdID, input.UserID, input.RefundID, amount, sourceBonus, rechargePoints, input.RequestFingerprint); err != nil {
			return err
		}
		if sourceBonus > 0 {
			if _, err := exec.ExecContext(txCtx, `
				UPDATE wallet_bonus_grants
				SET remaining_amount = remaining_amount - $1,
					frozen_amount = frozen_amount + $1,
					status = 'held', updated_at = NOW()
				WHERE id = $2
			`, sourceBonus, bonusGrantID); err != nil {
				return err
			}
			if _, err := exec.ExecContext(txCtx, `
				INSERT INTO wallet_hold_allocations (hold_id, bonus_grant_id, amount)
				VALUES ($1,$2,$3)
			`, holdID, bonusGrantID, sourceBonus); err != nil {
				return err
			}
		}
		if _, err := exec.ExecContext(txCtx, `
			UPDATE users SET
				balance = balance - $1, bonus_balance = bonus_balance - $2,
				frozen_balance = frozen_balance + $1, frozen_bonus_balance = frozen_bonus_balance + $2,
				updated_at = NOW()
			WHERE id = $3
		`, amount, sourceBonus, input.UserID); err != nil {
			return err
		}
		row.balance = walletMoney(row.balance - amount)
		row.bonus = walletMoney(row.bonus - sourceBonus)
		row.frozen = walletMoney(row.frozen + amount)
		row.frozenBonus = walletMoney(row.frozenBonus + sourceBonus)
		if err := insertWalletTransaction(txCtx, exec, input.UserID, "hold", -amount, -sourceBonus, -rechargePoints, amount,
			before, row, "payment_refund", input.RefundID, "wallet-hold:payment_refund:"+input.RefundID, input.Notes); err != nil {
			return err
		}
		result.Applied = true
		result.HoldID = holdID
		result.Status = "held"
		result.BonusAmount = sourceBonus
		result.RechargeAmount = rechargePoints
		result.Summary, err = loadWalletSummary(txCtx, exec, input.UserID)
		return err
	})
	return result, err
}

type lockedWalletHold struct {
	id, userID, purpose, referenceID, status                           string
	amount, bonus, recharge, refunded, refundedBonus, refundedRecharge float64
}

func loadWalletHoldForUpdate(ctx context.Context, exec sqlQueryExecutor, holdID string) (lockedWalletHold, error) {
	rows, err := exec.QueryContext(ctx, `
		SELECT id, user_id, purpose, reference_id, status, amount, bonus_amount, recharge_amount,
			refunded_amount, refunded_bonus_amount, refunded_recharge_amount
		FROM wallet_holds WHERE id = $1 FOR UPDATE
	`, holdID)
	if err != nil {
		return lockedWalletHold{}, err
	}
	defer rows.Close() //nolint:errcheck
	if !rows.Next() {
		return lockedWalletHold{}, service.ErrWalletHoldNotFound
	}
	var hold lockedWalletHold
	if err := rows.Scan(&hold.id, &hold.userID, &hold.purpose, &hold.referenceID, &hold.status, &hold.amount, &hold.bonus, &hold.recharge, &hold.refunded, &hold.refundedBonus, &hold.refundedRecharge); err != nil {
		return lockedWalletHold{}, err
	}
	return hold, rows.Err()
}

func walletHoldResult(hold lockedWalletHold, summary service.WalletSummary, applied bool) service.WalletHoldResult {
	return service.WalletHoldResult{Applied: applied, HoldID: hold.id, Status: hold.status, Amount: hold.amount, BonusAmount: hold.bonus, RechargeAmount: hold.recharge, Summary: summary}
}

func (r *userRepository) CaptureWalletHold(ctx context.Context, holdID, idempotencyKey string) (service.WalletHoldResult, error) {
	var result service.WalletHoldResult
	err := r.withWalletTx(ctx, func(txCtx context.Context, exec sqlQueryExecutor) error {
		hold, err := loadWalletHoldForUpdate(txCtx, exec, holdID)
		if err != nil {
			return err
		}
		if hold.status == "captured" || hold.status == "refunded" || hold.status == "partially_refunded" {
			summary, err := loadWalletSummary(txCtx, exec, hold.userID)
			result = walletHoldResult(hold, summary, false)
			return err
		}
		if hold.status != "held" {
			return service.ErrWalletHoldState
		}
		row, err := lockWalletUser(txCtx, exec, hold.userID)
		if err != nil {
			return err
		}
		if _, err := expireUserBonusTx(txCtx, exec, hold.userID, &row); err != nil {
			return err
		}
		before := row
		if _, err := exec.ExecContext(txCtx, `
			UPDATE wallet_bonus_grants g SET
				frozen_amount = g.frozen_amount - a.amount,
				spent_amount = g.spent_amount + CASE WHEN $2 NOT IN ('payment_refund', 'affiliate_refund') THEN a.amount ELSE 0 END,
				reversed_amount = g.reversed_amount + CASE WHEN $2 IN ('payment_refund', 'affiliate_refund') THEN a.amount ELSE 0 END,
				status = CASE
					WHEN $2 IN ('payment_refund', 'affiliate_refund') AND g.frozen_amount - a.amount = 0 AND g.remaining_amount = 0 THEN 'reversed'
					WHEN g.frozen_amount - a.amount = 0 AND g.remaining_amount = 0 THEN 'spent'
					ELSE g.status END,
				updated_at = NOW()
			FROM wallet_hold_allocations a WHERE a.hold_id = $1 AND a.bonus_grant_id = g.id
		`, hold.id, hold.purpose); err != nil {
			return err
		}
		if _, err := exec.ExecContext(txCtx, `
			UPDATE users SET frozen_balance = frozen_balance - $1,
				frozen_bonus_balance = frozen_bonus_balance - $2, updated_at = NOW()
			WHERE id = $3 AND frozen_balance >= $1 AND frozen_bonus_balance >= $2
		`, hold.amount, hold.bonus, hold.userID); err != nil {
			return err
		}
		if _, err := exec.ExecContext(txCtx, `UPDATE wallet_holds SET status = 'captured', updated_at = NOW() WHERE id = $1`, hold.id); err != nil {
			return err
		}
		row.frozen = walletMoney(row.frozen - hold.amount)
		row.frozenBonus = walletMoney(row.frozenBonus - hold.bonus)
		key := walletIdempotencyKey("wallet-capture:"+hold.id, idempotencyKey)
		if err := insertWalletTransaction(txCtx, exec, hold.userID, "capture", -hold.amount, -hold.bonus, -hold.recharge, -hold.amount, before, row, hold.purpose, hold.referenceID, key, "capture wallet hold"); err != nil {
			return err
		}
		hold.status = "captured"
		summary, err := loadWalletSummary(txCtx, exec, hold.userID)
		result = walletHoldResult(hold, summary, true)
		return err
	})
	return result, err
}

func (r *userRepository) ReleaseWalletHold(ctx context.Context, holdID, idempotencyKey string) (service.WalletHoldResult, error) {
	var result service.WalletHoldResult
	err := r.withWalletTx(ctx, func(txCtx context.Context, exec sqlQueryExecutor) error {
		hold, err := loadWalletHoldForUpdate(txCtx, exec, holdID)
		if err != nil {
			return err
		}
		if hold.status == "released" {
			summary, err := loadWalletSummary(txCtx, exec, hold.userID)
			result = walletHoldResult(hold, summary, false)
			return err
		}
		if hold.status != "held" {
			return service.ErrWalletHoldState
		}
		row, err := lockWalletUser(txCtx, exec, hold.userID)
		if err != nil {
			return err
		}
		if _, err := expireUserBonusTx(txCtx, exec, hold.userID, &row); err != nil {
			return err
		}
		before := row
		allocRows, err := exec.QueryContext(txCtx, `
			SELECT a.bonus_grant_id, a.amount, g.expires_at > NOW()
			FROM wallet_hold_allocations a JOIN wallet_bonus_grants g ON g.id = a.bonus_grant_id
			WHERE a.hold_id = $1 FOR UPDATE OF a, g
		`, hold.id)
		if err != nil {
			return err
		}
		type allocation struct {
			grantID string
			amount  float64
			restore bool
		}
		allocations := make([]allocation, 0)
		for allocRows.Next() {
			var item allocation
			if err := allocRows.Scan(&item.grantID, &item.amount, &item.restore); err != nil {
				_ = allocRows.Close()
				return err
			}
			allocations = append(allocations, item)
		}
		if err := allocRows.Close(); err != nil {
			return err
		}
		var restoredBonus float64
		for _, item := range allocations {
			if item.restore {
				restoredBonus = walletMoney(restoredBonus + item.amount)
			}
			if _, err := exec.ExecContext(txCtx, `
				UPDATE wallet_bonus_grants SET
					frozen_amount = frozen_amount - $1,
					remaining_amount = remaining_amount + $2,
					expired_amount = expired_amount + CASE WHEN $2 = 0 THEN $1 ELSE 0 END,
					status = CASE
						WHEN $2 > 0 AND frozen_amount - $1 > 0 THEN 'held'
						WHEN $2 > 0 THEN 'active'
						WHEN frozen_amount - $1 = 0 THEN 'expired'
						ELSE 'held_expired'
					END,
					updated_at = NOW()
				WHERE id = $3
			`, item.amount, func() float64 {
				if item.restore {
					return item.amount
				}
				return 0
			}(), item.grantID); err != nil {
				return err
			}
		}
		restored := walletMoney(hold.recharge + restoredBonus)
		if _, err := exec.ExecContext(txCtx, `
			UPDATE users SET
				balance = balance + $1, bonus_balance = bonus_balance + $2,
				frozen_balance = frozen_balance - $3, frozen_bonus_balance = frozen_bonus_balance - $4,
				updated_at = NOW()
			WHERE id = $5
		`, restored, restoredBonus, hold.amount, hold.bonus, hold.userID); err != nil {
			return err
		}
		if _, err := exec.ExecContext(txCtx, `UPDATE wallet_holds SET status = 'released', updated_at = NOW() WHERE id = $1`, hold.id); err != nil {
			return err
		}
		row.balance = walletMoney(row.balance + restored)
		row.bonus = walletMoney(row.bonus + restoredBonus)
		row.frozen = walletMoney(row.frozen - hold.amount)
		row.frozenBonus = walletMoney(row.frozenBonus - hold.bonus)
		key := walletIdempotencyKey("wallet-release:"+hold.id, idempotencyKey)
		if err := insertWalletTransaction(txCtx, exec, hold.userID, "release", restored, restoredBonus, hold.recharge, -hold.amount, before, row, hold.purpose, hold.referenceID, key, "release wallet hold"); err != nil {
			return err
		}
		hold.status = "released"
		summary, err := loadWalletSummary(txCtx, exec, hold.userID)
		result = walletHoldResult(hold, summary, true)
		return err
	})
	return result, err
}

func (r *userRepository) RefundWalletHold(ctx context.Context, holdID string, amount float64, idempotencyKey string) (service.WalletMutationResult, error) {
	amount = walletMoney(amount)
	result := service.WalletMutationResult{Amount: amount}
	err := r.withWalletTx(ctx, func(txCtx context.Context, exec sqlQueryExecutor) error {
		key := walletIdempotencyKey("wallet-refund:"+holdID, idempotencyKey)
		exists, err := walletTransactionExists(txCtx, exec, key)
		if err != nil {
			return err
		}
		hold, err := loadWalletHoldForUpdate(txCtx, exec, holdID)
		if err != nil {
			return err
		}
		if exists {
			result.Summary, err = loadWalletSummary(txCtx, exec, hold.userID)
			return err
		}
		if hold.status != "captured" && hold.status != "partially_refunded" {
			return service.ErrWalletHoldState
		}
		remaining := walletMoney(hold.amount - hold.refunded)
		if amount <= 0 || amount > remaining {
			amount = remaining
			result.Amount = amount
		}
		if amount <= 0 {
			result.Summary, err = loadWalletSummary(txCtx, exec, hold.userID)
			return err
		}
		row, err := lockWalletUser(txCtx, exec, hold.userID)
		if err != nil {
			return err
		}
		if _, err := expireUserBonusTx(txCtx, exec, hold.userID, &row); err != nil {
			return err
		}
		before := row
		remainingBonus := walletMoney(hold.bonus - hold.refundedBonus)
		remainingRecharge := walletMoney(hold.recharge - hold.refundedRecharge)
		bonusShare := 0.0
		if math.Abs(amount-remaining) < 0.00000001 {
			bonusShare = remainingBonus
		} else if remaining > 0 {
			bonusShare = walletMoney(amount * remainingBonus / remaining)
			bonusShare = math.Min(bonusShare, remainingBonus)
		}
		rechargeShare := walletMoney(amount - bonusShare)
		if rechargeShare > remainingRecharge {
			rechargeShare = remainingRecharge
			bonusShare = walletMoney(amount - rechargeShare)
		}
		allocRows, err := exec.QueryContext(txCtx, `
			SELECT a.id, a.bonus_grant_id, a.amount - a.refunded_amount, g.expires_at > NOW()
			FROM wallet_hold_allocations a JOIN wallet_bonus_grants g ON g.id = a.bonus_grant_id
			WHERE a.hold_id = $1 AND a.amount > a.refunded_amount
			ORDER BY g.expires_at ASC, a.created_at ASC FOR UPDATE OF a, g
		`, hold.id)
		if err != nil {
			return err
		}
		type refundableAllocation struct {
			id, grantID string
			remaining   float64
			restore     bool
		}
		allocations := make([]refundableAllocation, 0)
		for allocRows.Next() {
			var item refundableAllocation
			if err := allocRows.Scan(&item.id, &item.grantID, &item.remaining, &item.restore); err != nil {
				_ = allocRows.Close()
				return err
			}
			allocations = append(allocations, item)
		}
		if err := allocRows.Close(); err != nil {
			return err
		}
		toRefund := bonusShare
		var restoredBonus float64
		for _, item := range allocations {
			if toRefund <= 0 {
				break
			}
			part := walletMoney(math.Min(item.remaining, toRefund))
			if _, err := exec.ExecContext(txCtx, `UPDATE wallet_hold_allocations SET refunded_amount = refunded_amount + $1 WHERE id = $2`, part, item.id); err != nil {
				return err
			}
			if _, err := exec.ExecContext(txCtx, `
				UPDATE wallet_bonus_grants SET
					spent_amount = GREATEST(spent_amount - CASE WHEN $3 NOT IN ('payment_refund', 'affiliate_refund') THEN $1 ELSE 0 END, 0),
					reversed_amount = GREATEST(reversed_amount - CASE WHEN $3 IN ('payment_refund', 'affiliate_refund') THEN $1 ELSE 0 END, 0),
					remaining_amount = remaining_amount + CASE WHEN $4 THEN $1 ELSE 0 END,
					expired_amount = expired_amount + CASE WHEN $4 THEN 0 ELSE $1 END,
					status = CASE
						WHEN $4 AND frozen_amount > 0 THEN 'held'
						WHEN $4 THEN 'active'
						WHEN frozen_amount > 0 THEN 'held_expired'
						ELSE 'expired'
					END,
					updated_at = NOW()
				WHERE id = $2
			`, part, item.grantID, hold.purpose, item.restore); err != nil {
				return err
			}
			if item.restore {
				restoredBonus = walletMoney(restoredBonus + part)
			}
			toRefund = walletMoney(toRefund - part)
		}
		restored := walletMoney(rechargeShare + restoredBonus)
		if _, err := exec.ExecContext(txCtx, `
			UPDATE users SET balance = balance + $1, bonus_balance = bonus_balance + $2, updated_at = NOW() WHERE id = $3
		`, restored, restoredBonus, hold.userID); err != nil {
			return err
		}
		newRefunded := walletMoney(hold.refunded + amount)
		status := "partially_refunded"
		if newRefunded >= hold.amount-0.00000001 {
			status = "refunded"
		}
		if _, err := exec.ExecContext(txCtx, `
			UPDATE wallet_holds SET refunded_amount = refunded_amount + $1,
				refunded_bonus_amount = refunded_bonus_amount + $2,
				refunded_recharge_amount = refunded_recharge_amount + $3,
				status = $4, updated_at = NOW() WHERE id = $5
		`, amount, bonusShare, rechargeShare, status, hold.id); err != nil {
			return err
		}
		row.balance = walletMoney(row.balance + restored)
		row.bonus = walletMoney(row.bonus + restoredBonus)
		result.BonusAmount = restoredBonus
		result.RechargeAmount = rechargeShare
		if err := insertWalletTransaction(txCtx, exec, hold.userID, "refund", restored, restoredBonus, rechargeShare, 0, before, row, hold.purpose, hold.referenceID, key, "restore subscription wallet payment"); err != nil {
			return err
		}
		result.Applied = true
		result.Summary, err = loadWalletSummary(txCtx, exec, hold.userID)
		return err
	})
	return result, err
}

func (r *userRepository) ListWalletTransactions(ctx context.Context, userID string, page, pageSize int) (service.WalletTransactionPage, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	exec := txAwareSQLExecutor(ctx, r.sql, r.client)
	if exec == nil {
		return service.WalletTransactionPage{}, errors.New("wallet SQL executor is not configured")
	}
	var total int64
	countRows, err := exec.QueryContext(ctx, `SELECT COUNT(*) FROM wallet_transactions WHERE user_id = $1`, userID)
	if err != nil {
		return service.WalletTransactionPage{}, err
	}
	if countRows.Next() {
		err = countRows.Scan(&total)
	}
	_ = countRows.Close()
	if err != nil {
		return service.WalletTransactionPage{}, err
	}
	rows, err := exec.QueryContext(ctx, `
		SELECT id, action, amount, bonus_amount, recharge_amount, frozen_amount,
			balance_before, balance_after, bonus_before, bonus_after,
			source_type, source_id, notes, created_at
		FROM wallet_transactions WHERE user_id = $1
		ORDER BY created_at DESC, id DESC LIMIT $2 OFFSET $3
	`, userID, pageSize, (page-1)*pageSize)
	if err != nil {
		return service.WalletTransactionPage{}, err
	}
	defer rows.Close() //nolint:errcheck
	items := make([]service.WalletTransaction, 0)
	for rows.Next() {
		var item service.WalletTransaction
		if err := rows.Scan(&item.ID, &item.Action, &item.Amount, &item.BonusAmount, &item.RechargeAmount, &item.FrozenAmount,
			&item.BalanceBefore, &item.BalanceAfter, &item.BonusBefore, &item.BonusAfter,
			&item.SourceType, &item.SourceID, &item.Notes, &item.CreatedAt); err != nil {
			return service.WalletTransactionPage{}, err
		}
		items = append(items, item)
	}
	return service.WalletTransactionPage{Items: items, Total: total}, rows.Err()
}

func (r *userRepository) ExpireBonusBalances(ctx context.Context, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 200
	}
	exec := txAwareSQLExecutor(ctx, r.sql, r.client)
	if exec == nil {
		return nil, errors.New("wallet SQL executor is not configured")
	}
	rows, err := exec.QueryContext(ctx, `
		SELECT DISTINCT user_id FROM wallet_bonus_grants
		WHERE remaining_amount > 0 AND expires_at <= NOW()
		ORDER BY user_id LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	userIDs := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			_ = rows.Close()
			return nil, err
		}
		userIDs = append(userIDs, id)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	processed := make([]string, 0, len(userIDs))
	for _, userID := range userIDs {
		err := r.withWalletTx(ctx, func(txCtx context.Context, txExec sqlQueryExecutor) error {
			row, err := lockWalletUser(txCtx, txExec, userID)
			if err != nil {
				return err
			}
			expired, err := expireUserBonusTx(txCtx, txExec, userID, &row)
			if expired > 0 {
				processed = append(processed, userID)
			}
			return err
		})
		if err != nil {
			return processed, err
		}
	}
	return processed, nil
}

var _ service.WalletRepository = (*userRepository)(nil)
