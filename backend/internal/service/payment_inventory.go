package service

import (
	"context"
	"fmt"
	"time"

	dbent "github.com/AsukaCC/EasySub2api/ent"
	"github.com/AsukaCC/EasySub2api/ent/paymentorder"
	"github.com/AsukaCC/EasySub2api/ent/subscriptionplan"
	entuser "github.com/AsukaCC/EasySub2api/ent/user"
	infraerrors "github.com/AsukaCC/EasySub2api/internal/pkg/errors"
)

const (
	InventoryStatusNone     = "NONE"
	InventoryStatusReserved = "RESERVED"
	InventoryStatusConsumed = "CONSUMED"
	InventoryStatusReleased = "RELEASED"
)

func reserveSubscriptionInventory(ctx context.Context, tx *dbent.Tx, planID string) (*dbent.SubscriptionPlan, bool, error) {
	plan, err := tx.SubscriptionPlan.Query().Where(subscriptionplan.IDEQ(planID)).ForUpdate().Only(ctx)
	if err != nil || !plan.ForSale {
		return nil, false, infraerrors.NotFound("PLAN_NOT_AVAILABLE", "plan not found or not for sale")
	}
	if plan.StockQuantity == nil {
		return plan, false, nil
	}
	if plan.StockFrozen >= *plan.StockQuantity {
		return nil, false, infraerrors.Conflict("PLAN_SOLD_OUT", "subscription plan is sold out")
	}
	plan, err = tx.SubscriptionPlan.UpdateOneID(plan.ID).AddStockFrozen(1).Save(ctx)
	if err != nil {
		return nil, false, fmt.Errorf("freeze subscription inventory: %w", err)
	}
	return plan, true, nil
}

func (s *PaymentService) consumeSubscriptionInventoryTx(ctx context.Context, client *dbent.Client, orderID string) error {
	preview, err := client.PaymentOrder.Query().Where(paymentorder.IDEQ(orderID)).Only(ctx)
	if err != nil {
		return fmt.Errorf("load order inventory: %w", err)
	}
	if _, err := client.User.Query().Where(entuser.IDEQ(preview.UserID)).ForUpdate().Only(ctx); err != nil {
		return fmt.Errorf("lock inventory user: %w", err)
	}
	if preview.PlanID != nil {
		if _, err := client.SubscriptionPlan.Query().Where(subscriptionplan.IDEQ(*preview.PlanID)).ForUpdate().Only(ctx); err != nil {
			return infraerrors.Conflict("PLAN_NOT_AVAILABLE", "subscription plan no longer exists")
		}
	}
	order, err := client.PaymentOrder.Query().Where(paymentorder.IDEQ(orderID)).ForUpdate().Only(ctx)
	if err != nil {
		return fmt.Errorf("lock order inventory: %w", err)
	}
	if order.InventoryStatus == InventoryStatusConsumed || order.InventoryStatus == InventoryStatusNone {
		return nil
	}
	if order.PlanID == nil {
		return infraerrors.Conflict("PLAN_NOT_AVAILABLE", "subscription order has no plan")
	}
	plan, err := client.SubscriptionPlan.Query().Where(subscriptionplan.IDEQ(*order.PlanID)).Only(ctx)
	if err != nil {
		return infraerrors.Conflict("PLAN_NOT_AVAILABLE", "subscription plan no longer exists")
	}
	if order.InventoryStatus == InventoryStatusReleased {
		if plan.StockQuantity == nil {
			_, err = client.PaymentOrder.UpdateOneID(order.ID).
				SetInventoryStatus(InventoryStatusConsumed).SetInventoryConsumedAt(time.Now().UTC()).Save(ctx)
			return err
		}
		if plan.StockFrozen >= *plan.StockQuantity {
			return infraerrors.Conflict("PLAN_SOLD_OUT", "subscription plan is sold out")
		}
		plan, err = client.SubscriptionPlan.UpdateOneID(plan.ID).AddStockFrozen(1).Save(ctx)
		if err != nil {
			return fmt.Errorf("reacquire subscription inventory: %w", err)
		}
	}
	if plan.StockFrozen <= 0 {
		return infraerrors.Conflict("PLAN_INVENTORY_CONFLICT", "frozen subscription inventory is missing")
	}
	update := client.SubscriptionPlan.UpdateOneID(plan.ID).AddStockFrozen(-1)
	if plan.StockQuantity != nil {
		if *plan.StockQuantity <= 0 {
			return infraerrors.Conflict("PLAN_SOLD_OUT", "subscription plan is sold out")
		}
		update.AddStockQuantity(-1)
	}
	if _, err := update.Save(ctx); err != nil {
		return fmt.Errorf("consume subscription inventory: %w", err)
	}
	now := time.Now().UTC()
	if _, err := client.PaymentOrder.UpdateOneID(order.ID).
		SetInventoryStatus(InventoryStatusConsumed).
		SetInventoryConsumedAt(now).
		ClearInventoryReleasedAt().
		Save(ctx); err != nil {
		return fmt.Errorf("mark subscription inventory consumed: %w", err)
	}
	return nil
}

func (s *PaymentService) releaseSubscriptionInventory(ctx context.Context, orderID string) error {
	preview, err := s.entClient.PaymentOrder.Query().Where(paymentorder.IDEQ(orderID)).Only(ctx)
	if err != nil {
		return fmt.Errorf("load order inventory for release: %w", err)
	}
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin inventory release transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)
	if _, err := tx.User.Query().Where(entuser.IDEQ(preview.UserID)).ForUpdate().Only(txCtx); err != nil {
		return fmt.Errorf("lock inventory user for release: %w", err)
	}
	if preview.PlanID != nil {
		if _, err := tx.SubscriptionPlan.Query().Where(subscriptionplan.IDEQ(*preview.PlanID)).ForUpdate().Only(txCtx); err != nil {
			return fmt.Errorf("lock plan inventory for release: %w", err)
		}
	}
	order, err := tx.PaymentOrder.Query().Where(paymentorder.IDEQ(orderID)).ForUpdate().Only(txCtx)
	if err != nil {
		return fmt.Errorf("lock order inventory for release: %w", err)
	}
	if order.InventoryStatus != InventoryStatusReserved {
		return nil
	}
	if order.PlanID == nil {
		return nil
	}
	plan, err := tx.SubscriptionPlan.Query().Where(subscriptionplan.IDEQ(*order.PlanID)).Only(txCtx)
	if err != nil {
		return fmt.Errorf("lock plan inventory for release: %w", err)
	}
	if plan.StockFrozen <= 0 {
		return infraerrors.Conflict("PLAN_INVENTORY_CONFLICT", "frozen subscription inventory is missing")
	}
	if _, err := tx.SubscriptionPlan.UpdateOneID(plan.ID).AddStockFrozen(-1).Save(txCtx); err != nil {
		return fmt.Errorf("release plan inventory: %w", err)
	}
	if _, err := tx.PaymentOrder.UpdateOneID(order.ID).
		SetInventoryStatus(InventoryStatusReleased).
		SetInventoryReleasedAt(time.Now().UTC()).
		Save(txCtx); err != nil {
		return fmt.Errorf("mark plan inventory released: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit inventory release: %w", err)
	}
	return nil
}
