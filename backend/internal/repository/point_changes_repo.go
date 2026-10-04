package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/AsukaCC/EasySub2api/internal/service"
)

// Capture changes only the frozen bucket: its legacy amount must not be billed twice.
const pointChangeRows = `
 FROM wallet_transactions t
 JOIN users u ON u.id = t.user_id
 CROSS JOIN LATERAL (VALUES
   ('recharge', t.recharge_amount, t.balance_before, t.balance_after),
   ('bonus', t.bonus_amount, t.bonus_before, t.bonus_after)
 ) p(kind, raw_amount, balance_before, balance_after)
 CROSS JOIN LATERAL (SELECT CASE WHEN t.action = 'capture' THEN 0 ELSE p.raw_amount END AS delta) d
 WHERE p.raw_amount <> 0`

func pointChangeWhere(f service.PointChangeFilter) (string, []any) {
	where := pointChangeRows
	args := []any{}
	add := func(clause string, value any) {
		args = append(args, value)
		where += fmt.Sprintf(clause, len(args))
	}
	if f.Keyword != "" {
		// Treat percent and underscore as literal search characters.
		keyword := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(strings.TrimSpace(f.Keyword))
		add(" AND (u.email ILIKE $%[1]d OR u.username ILIKE $%[1]d)", "%"+keyword+"%")
	}
	if f.UserID != "" {
		add(" AND t.user_id = $%d::uuid", f.UserID)
	}
	if f.PointType != "" {
		add(" AND p.kind = $%d", f.PointType)
	}
	if f.Direction == "increase" {
		where += " AND d.delta > 0"
	}
	if f.Direction == "decrease" {
		where += " AND d.delta < 0"
	}
	if f.StartTime != nil {
		add(" AND t.created_at >= $%d", *f.StartTime)
	}
	if f.EndTime != nil {
		add(" AND t.created_at < $%d", *f.EndTime)
	}
	return where, args
}

func (r *userRepository) ListPointChanges(ctx context.Context, f service.PointChangeFilter) (service.PointChangePage, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 {
		f.PageSize = 20
	}
	if f.PageSize > 100 {
		f.PageSize = 100
	}
	result := service.PointChangePage{Items: []service.PointChange{}, Page: f.Page, PageSize: f.PageSize}
	exec := txAwareSQLExecutor(ctx, r.sql, r.client)
	if exec == nil {
		return result, fmt.Errorf("wallet SQL executor is not configured")
	}
	where, args := pointChangeWhere(f)
	rows, err := exec.QueryContext(ctx, "SELECT COUNT(*)"+where, args...)
	if err != nil {
		return result, err
	}
	if rows.Next() {
		err = rows.Scan(&result.Total)
	}
	if err == nil {
		err = rows.Err()
	}
	_ = rows.Close()
	if err != nil {
		return result, err
	}
	query := `SELECT t.id, t.user_id, u.email, COALESCE(u.username, ''), p.kind, d.delta,
 CASE WHEN t.action = 'hold' THEN -p.raw_amount
      WHEN t.action = 'capture' THEN p.raw_amount
      WHEN t.action = 'release' THEN -p.raw_amount ELSE 0 END,
 CASE WHEN t.snapshot_version > 0 AND p.balance_after - p.balance_before = d.delta THEN t.balance_before + t.bonus_before END,
 CASE WHEN t.snapshot_version > 0 AND p.balance_after - p.balance_before = d.delta THEN t.balance_after + t.bonus_after END,
 t.action, t.source_type, t.source_id, t.notes, t.created_at` + where +
		fmt.Sprintf(" ORDER BY t.created_at DESC, t.id DESC, p.kind ASC LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	args = append(args, f.PageSize, (f.Page-1)*f.PageSize)
	rows, err = exec.QueryContext(ctx, query, args...)
	if err != nil {
		return result, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var item service.PointChange
		if err := rows.Scan(&item.TransactionID, &item.UserID, &item.Email, &item.Username,
			&item.PointType, &item.Amount, &item.FrozenAmount, &item.BalanceBefore, &item.BalanceAfter,
			&item.Action, &item.SourceType, &item.SourceID, &item.Notes, &item.CreatedAt); err != nil {
			return result, err
		}
		item.ID = item.TransactionID + ":" + item.PointType
		result.Items = append(result.Items, item)
	}
	return result, rows.Err()
}

var _ service.PointChangeReader = (*userRepository)(nil)
