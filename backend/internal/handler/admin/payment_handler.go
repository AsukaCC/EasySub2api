package admin

import (
	"strconv"
	"time"

	dbent "github.com/AsukaCC/EasySub2api/ent"
	"github.com/AsukaCC/EasySub2api/internal/pkg/response"
	"github.com/AsukaCC/EasySub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// PaymentHandler handles admin payment management.
type PaymentHandler struct {
	paymentService *service.PaymentService
	configService  *service.PaymentConfigService
}

// NewPaymentHandler creates a new admin PaymentHandler.
func NewPaymentHandler(paymentService *service.PaymentService, configService *service.PaymentConfigService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
		configService:  configService,
	}
}

// --- Dashboard ---

// GetDashboard returns payment dashboard statistics.
// GET /api/v1/admin/payment/dashboard
func (h *PaymentHandler) GetDashboard(c *gin.Context) {
	days := 30
	if d := c.Query("days"); d != "" {
		if v, err := strconv.Atoi(d); err == nil && v > 0 {
			days = v
		}
	}
	stats, err := h.paymentService.GetDashboardStats(c.Request.Context(), days)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, stats)
}

// --- Orders ---

// ListOrders returns a paginated list of all payment orders.
// GET /api/v1/admin/payment/orders
func (h *PaymentHandler) ListOrders(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	var userID string
	if uid := c.Query("user_id"); uid != "" {
		if v, err := parseEntityID(uid); err == nil {
			userID = v
		}
	}
	orders, total, err := h.paymentService.AdminListOrders(c.Request.Context(), userID, service.OrderListParams{
		Page:        page,
		PageSize:    pageSize,
		Status:      c.Query("status"),
		OrderType:   c.Query("order_type"),
		PaymentType: c.Query("payment_type"),
		Keyword:     c.Query("keyword"),
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, sanitizeAdminPaymentOrdersForResponse(orders), int64(total), page, pageSize)
}

// GetOrderDetail returns detailed information about a single order.
// GET /api/v1/admin/payment/orders/:id
func (h *PaymentHandler) GetOrderDetail(c *gin.Context) {
	orderID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	order, err := h.paymentService.GetOrderByID(c.Request.Context(), orderID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	auditLogs, _ := h.paymentService.GetOrderAuditLogs(c.Request.Context(), orderID)
	response.Success(c, gin.H{"order": sanitizeAdminPaymentOrderForResponse(order), "auditLogs": auditLogs})
}

// CancelOrder cancels a pending order (admin).
// POST /api/v1/admin/payment/orders/:id/cancel
func (h *PaymentHandler) CancelOrder(c *gin.Context) {
	orderID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	msg, err := h.paymentService.AdminCancelOrder(c.Request.Context(), orderID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": msg})
}

// RetryFulfillment retries fulfillment for a paid order.
// POST /api/v1/admin/payment/orders/:id/retry
func (h *PaymentHandler) RetryFulfillment(c *gin.Context) {
	orderID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.paymentService.RetryFulfillment(c.Request.Context(), orderID); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "fulfillment retried"})
}

type AdminPaymentOrderResult struct {
	ID                      string         `json:"id"`
	UserID                  string         `json:"user_id"`
	UserEmail               string         `json:"user_email,omitempty"`
	UserName                string         `json:"user_name,omitempty"`
	UserNotes               *string        `json:"user_notes,omitempty"`
	Amount                  float64        `json:"amount"`
	PayAmount               float64        `json:"pay_amount"`
	WalletAmount            float64        `json:"wallet_amount"`
	WalletBonusAmount       float64        `json:"wallet_bonus_amount"`
	WalletRechargeAmount    float64        `json:"wallet_recharge_amount"`
	GatewayBaseAmount       float64        `json:"gateway_base_amount"`
	WalletOnly              bool           `json:"wallet_only"`
	FeeRate                 float64        `json:"fee_rate"`
	PrincipalAmount         float64        `json:"principal_amount"`
	FeeAmount               float64        `json:"fee_amount"`
	BasePoints              float64        `json:"base_points"`
	BonusPoints             float64        `json:"bonus_points"`
	CreditedPoints          float64        `json:"credited_points"`
	BonusTierSnapshot       map[string]any `json:"bonus_tier_snapshot,omitempty"`
	BonusExpiresAt          *time.Time     `json:"bonus_expires_at,omitempty"`
	BonusGrantID            *string        `json:"bonus_grant_id,omitempty"`
	AffiliateRebatePoints   float64        `json:"affiliate_rebate_points"`
	RefundDeadline          *time.Time     `json:"refund_deadline,omitempty"`
	Currency                string         `json:"currency"`
	RechargeCode            string         `json:"recharge_code,omitempty"`
	OutTradeNo              string         `json:"out_trade_no"`
	PaymentType             string         `json:"payment_type"`
	PaymentTradeNo          string         `json:"payment_trade_no,omitempty"`
	PayURL                  *string        `json:"pay_url,omitempty"`
	QRCode                  *string        `json:"qr_code,omitempty"`
	QRCodeImg               *string        `json:"qr_code_img,omitempty"`
	OrderType               string         `json:"order_type"`
	PlanID                  *string        `json:"plan_id,omitempty"`
	SubscriptionGroupID     *string        `json:"subscription_group_id,omitempty"`
	SubscriptionDays        *int           `json:"subscription_days,omitempty"`
	ProviderInstanceID      *string        `json:"provider_instance_id,omitempty"`
	ProviderKey             *string        `json:"provider_key,omitempty"`
	Status                  string         `json:"status"`
	RefundAmount            float64        `json:"refund_amount"`
	RefundedPrincipalAmount float64        `json:"refunded_principal_amount"`
	RefundedFeeAmount       float64        `json:"refunded_fee_amount"`
	RefundedGatewayAmount   float64        `json:"refunded_gateway_amount"`
	ReversedBasePoints      float64        `json:"reversed_base_points"`
	ReversedBonusPoints     float64        `json:"reversed_bonus_points"`
	ReversedAffiliatePoints float64        `json:"reversed_affiliate_points"`
	RefundReason            *string        `json:"refund_reason,omitempty"`
	RefundAt                *time.Time     `json:"refund_at,omitempty"`
	ForceRefund             bool           `json:"force_refund,omitempty"`
	RefundRequestedAt       *time.Time     `json:"refund_requested_at,omitempty"`
	RefundRequestReason     *string        `json:"refund_request_reason,omitempty"`
	RefundRequestedBy       *string        `json:"refund_requested_by,omitempty"`
	ExpiresAt               time.Time      `json:"expires_at"`
	PaidAt                  *time.Time     `json:"paid_at,omitempty"`
	CompletedAt             *time.Time     `json:"completed_at,omitempty"`
	FailedAt                *time.Time     `json:"failed_at,omitempty"`
	FailedReason            *string        `json:"failed_reason,omitempty"`
	ClientIP                string         `json:"client_ip,omitempty"`
	SrcHost                 string         `json:"src_host,omitempty"`
	SrcURL                  *string        `json:"src_url,omitempty"`
	CreatedAt               time.Time      `json:"created_at"`
	UpdatedAt               time.Time      `json:"updated_at"`
}

func sanitizeAdminPaymentOrdersForResponse(orders []*dbent.PaymentOrder) []*AdminPaymentOrderResult {
	out := make([]*AdminPaymentOrderResult, 0, len(orders))
	for _, order := range orders {
		if item := sanitizeAdminPaymentOrderForResponse(order); item != nil {
			out = append(out, item)
		}
	}
	return out
}

func sanitizeAdminPaymentOrderForResponse(order *dbent.PaymentOrder) *AdminPaymentOrderResult {
	if order == nil {
		return nil
	}
	return &AdminPaymentOrderResult{
		ID:                      order.ID,
		UserID:                  order.UserID,
		UserEmail:               order.UserEmail,
		UserName:                order.UserName,
		UserNotes:               order.UserNotes,
		Amount:                  order.Amount,
		PayAmount:               order.PayAmount,
		WalletAmount:            order.WalletAmount,
		WalletBonusAmount:       order.WalletBonusAmount,
		WalletRechargeAmount:    order.WalletRechargeAmount,
		GatewayBaseAmount:       order.GatewayBaseAmount,
		WalletOnly:              order.WalletOnly,
		FeeRate:                 order.FeeRate,
		PrincipalAmount:         order.PrincipalAmount,
		FeeAmount:               order.FeeAmount,
		BasePoints:              order.BasePoints,
		BonusPoints:             order.BonusPoints,
		CreditedPoints:          order.CreditedPoints,
		BonusTierSnapshot:       order.BonusTierSnapshot,
		BonusExpiresAt:          order.BonusExpiresAt,
		BonusGrantID:            order.BonusGrantID,
		AffiliateRebatePoints:   order.AffiliateRebatePoints,
		RefundDeadline:          order.RefundDeadline,
		Currency:                service.PaymentOrderCurrency(order),
		RechargeCode:            order.RechargeCode,
		OutTradeNo:              order.OutTradeNo,
		PaymentType:             order.PaymentType,
		PaymentTradeNo:          order.PaymentTradeNo,
		PayURL:                  order.PayURL,
		QRCode:                  order.QrCode,
		QRCodeImg:               order.QrCodeImg,
		OrderType:               order.OrderType,
		PlanID:                  order.PlanID,
		SubscriptionGroupID:     order.SubscriptionGroupID,
		SubscriptionDays:        order.SubscriptionDays,
		ProviderInstanceID:      order.ProviderInstanceID,
		ProviderKey:             order.ProviderKey,
		Status:                  order.Status,
		RefundAmount:            order.RefundAmount,
		RefundedPrincipalAmount: order.RefundedPrincipalAmount,
		RefundedFeeAmount:       order.RefundedFeeAmount,
		RefundedGatewayAmount:   order.RefundedGatewayAmount,
		ReversedBasePoints:      order.ReversedBasePoints,
		ReversedBonusPoints:     order.ReversedBonusPoints,
		ReversedAffiliatePoints: order.ReversedAffiliatePoints,
		RefundReason:            order.RefundReason,
		RefundAt:                order.RefundAt,
		ForceRefund:             order.ForceRefund,
		RefundRequestedAt:       order.RefundRequestedAt,
		RefundRequestReason:     order.RefundRequestReason,
		RefundRequestedBy:       order.RefundRequestedBy,
		ExpiresAt:               order.ExpiresAt,
		PaidAt:                  order.PaidAt,
		CompletedAt:             order.CompletedAt,
		FailedAt:                order.FailedAt,
		FailedReason:            order.FailedReason,
		ClientIP:                order.ClientIP,
		SrcHost:                 order.SrcHost,
		SrcURL:                  order.SrcURL,
		CreatedAt:               order.CreatedAt,
		UpdatedAt:               order.UpdatedAt,
	}
}

// AdminProcessRefundRequest is the request body for admin refund processing.
type AdminProcessRefundRequest struct {
	Amount          *float64 `json:"amount,omitempty"`
	PrincipalAmount *float64 `json:"principal_amount,omitempty"`
	Reason          string   `json:"reason"`
	Force           bool     `json:"force"`
	DeductBalance   bool     `json:"deduct_balance"`
}

type AdminReviewRefundTicketRequest struct {
	Decision                string   `json:"decision" binding:"required"`
	ApprovedPrincipalAmount *float64 `json:"approved_principal_amount,omitempty"`
	ReviewNote              string   `json:"review_note"`
	AffiliateAction         string   `json:"affiliate_action"`
}

// ProcessRefund processes a refund for an order (admin).
// POST /api/v1/admin/payment/orders/:id/refund
func (h *PaymentHandler) ProcessRefund(c *gin.Context) {
	orderID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req AdminProcessRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	if req.Amount != nil && req.PrincipalAmount != nil {
		response.BadRequest(c, "amount and principal_amount cannot be provided together")
		return
	}
	principal := req.PrincipalAmount
	if principal == nil {
		principal = req.Amount
	}
	result, err := h.paymentService.CreateAdminPaymentRefund(c.Request.Context(), service.CreatePaymentRefundInput{
		OrderID: orderID, RequestedBy: getAdminIDFromContext(c),
		IdempotencyKey: c.GetHeader("Idempotency-Key"), Principal: principal, Reason: req.Reason,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

// QueryAndFinalizeRefund queries the provider refund status and finalizes a pending refund.
// POST /api/v1/admin/payment/orders/:id/refund/query
func (h *PaymentHandler) QueryAndFinalizeRefund(c *gin.Context) {
	orderID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	result, err := h.paymentService.QueryPaymentRefund(c.Request.Context(), orderID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

// ListRefundTickets returns refund tickets for manual review.
// GET /api/v1/admin/payment/refund-tickets
func (h *PaymentHandler) ListRefundTickets(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	result, err := h.paymentService.AdminListRefundTickets(c.Request.Context(), c.Query("status"), page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, result.Items, int64(result.Total), page, pageSize)
}

// ReviewRefundTicket approves or rejects one refund ticket. Approved tickets
// use a stable ticket-scoped idempotency key for provider retries.
// POST /api/v1/admin/payment/refund-tickets/:id/review
func (h *PaymentHandler) ReviewRefundTicket(c *gin.Context) {
	ticketID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req AdminReviewRefundTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	result, err := h.paymentService.ReviewRefundTicket(c.Request.Context(), service.ReviewRefundTicketInput{
		TicketID: ticketID, ReviewerID: getAdminIDFromContext(c),
		Decision: req.Decision, ApprovedPrincipalAmount: req.ApprovedPrincipalAmount,
		ReviewNote: req.ReviewNote, AffiliateAction: req.AffiliateAction,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

// --- Subscription Plans ---

// ListPlans returns all subscription plans.
// GET /api/v1/admin/payment/plans
func (h *PaymentHandler) ListPlans(c *gin.Context) {
	plans, err := h.configService.ListPlans(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	groupInfo := h.configService.GetGroupInfoMap(c.Request.Context(), plans)
	response.Success(c, adminSubscriptionPlansForResponse(plans, groupInfo))
}

type AdminSubscriptionPlanResult struct {
	ID                  string    `json:"id"`
	GroupID             string    `json:"group_id"`
	GroupPlatform       string    `json:"group_platform,omitempty"`
	GroupName           string    `json:"group_name,omitempty"`
	RateMultiplier      float64   `json:"rate_multiplier,omitempty"`
	DailyLimitUSD       *float64  `json:"daily_limit_usd,omitempty"`
	DailyLimitPoints    *float64  `json:"daily_limit_points,omitempty"`
	WeeklyLimitUSD      *float64  `json:"weekly_limit_usd,omitempty"`
	WeeklyLimitPoints   *float64  `json:"weekly_limit_points,omitempty"`
	MonthlyLimitUSD     *float64  `json:"monthly_limit_usd,omitempty"`
	MonthlyLimitPoints  *float64  `json:"monthly_limit_points,omitempty"`
	Name                string    `json:"name"`
	Description         string    `json:"description"`
	Price               float64   `json:"price"`
	PricePoints         float64   `json:"price_points"`
	OriginalPrice       *float64  `json:"original_price,omitempty"`
	OriginalPricePoints *float64  `json:"original_price_points,omitempty"`
	Currency            string    `json:"currency,omitempty"`
	ValidityDays        int       `json:"validity_days"`
	ValidityUnit        string    `json:"validity_unit"`
	Features            string    `json:"features"`
	ProductName         string    `json:"product_name"`
	ForSale             bool      `json:"for_sale"`
	SortOrder           int       `json:"sort_order"`
	CreatedAt           time.Time `json:"created_at,omitempty"`
	UpdatedAt           time.Time `json:"updated_at,omitempty"`
}

func adminSubscriptionPlansForResponse(plans []*dbent.SubscriptionPlan, groupInfo map[string]service.PlanGroupInfo) []AdminSubscriptionPlanResult {
	result := make([]AdminSubscriptionPlanResult, 0, len(plans))
	for _, p := range plans {
		if p == nil {
			continue
		}
		gi := groupInfo[p.GroupID]
		result = append(result, AdminSubscriptionPlanResult{
			ID:                  p.ID,
			GroupID:             p.GroupID,
			GroupPlatform:       gi.Platform,
			GroupName:           gi.Name,
			RateMultiplier:      gi.RateMultiplier,
			DailyLimitUSD:       gi.DailyLimitUSD,
			DailyLimitPoints:    gi.DailyLimitUSD,
			WeeklyLimitUSD:      gi.WeeklyLimitUSD,
			WeeklyLimitPoints:   gi.WeeklyLimitUSD,
			MonthlyLimitUSD:     gi.MonthlyLimitUSD,
			MonthlyLimitPoints:  gi.MonthlyLimitUSD,
			Name:                p.Name,
			Description:         p.Description,
			Price:               p.Price,
			PricePoints:         p.Price,
			OriginalPrice:       p.OriginalPrice,
			OriginalPricePoints: p.OriginalPrice,
			Currency:            p.Currency,
			ValidityDays:        p.ValidityDays,
			ValidityUnit:        p.ValidityUnit,
			Features:            p.Features,
			ProductName:         p.ProductName,
			ForSale:             p.ForSale,
			SortOrder:           p.SortOrder,
			CreatedAt:           p.CreatedAt,
			UpdatedAt:           p.UpdatedAt,
		})
	}
	return result
}

// CreatePlan creates a new subscription plan.
// POST /api/v1/admin/payment/plans
func (h *PaymentHandler) CreatePlan(c *gin.Context) {
	var req service.CreatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	plan, err := h.configService.CreatePlan(c.Request.Context(), req)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, plan)
}

// UpdatePlan updates an existing subscription plan.
// PUT /api/v1/admin/payment/plans/:id
func (h *PaymentHandler) UpdatePlan(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req service.UpdatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	plan, err := h.configService.UpdatePlan(c.Request.Context(), id, req)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, plan)
}

// DeletePlan deletes a subscription plan.
// DELETE /api/v1/admin/payment/plans/:id
func (h *PaymentHandler) DeletePlan(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.configService.DeletePlan(c.Request.Context(), id); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "deleted"})
}

// --- Provider Instances ---

// ListProviders returns all payment provider instances.
// GET /api/v1/admin/payment/providers
func (h *PaymentHandler) ListProviders(c *gin.Context) {
	providers, err := h.configService.ListProviderInstancesWithConfig(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, providers)
}

// CreateProvider creates a new payment provider instance.
// POST /api/v1/admin/payment/providers
func (h *PaymentHandler) CreateProvider(c *gin.Context) {
	var req service.CreateProviderInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	inst, err := h.configService.CreateProviderInstance(c.Request.Context(), req)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	h.paymentService.RefreshProviders(c.Request.Context())
	response.Created(c, inst)
}

// UpdateProvider updates an existing payment provider instance.
// PUT /api/v1/admin/payment/providers/:id
func (h *PaymentHandler) UpdateProvider(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req service.UpdateProviderInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	inst, err := h.configService.UpdateProviderInstance(c.Request.Context(), id, req)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	h.paymentService.RefreshProviders(c.Request.Context())
	response.Success(c, inst)
}

// DeleteProvider deletes a payment provider instance.
// DELETE /api/v1/admin/payment/providers/:id
func (h *PaymentHandler) DeleteProvider(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.configService.DeleteProviderInstance(c.Request.Context(), id); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	h.paymentService.RefreshProviders(c.Request.Context())
	response.Success(c, gin.H{"message": "deleted"})
}

// parseIDParam parses a UUID path parameter.
// Returns the parsed ID and true on success; on failure it writes a BadRequest response and returns false.
func parseIDParam(c *gin.Context, paramName string) (string, bool) {
	id, err := parseEntityID(c.Param(paramName))
	if err != nil {
		response.BadRequest(c, "Invalid "+paramName)
		return "", false
	}
	return id, true
}

// --- Config ---

// GetConfig returns the payment configuration (admin view).
// GET /api/v1/admin/payment/config
func (h *PaymentHandler) GetConfig(c *gin.Context) {
	cfg, err := h.configService.GetPaymentConfig(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, cfg)
}

// UpdateConfig updates the payment configuration.
// PUT /api/v1/admin/payment/config
func (h *PaymentHandler) UpdateConfig(c *gin.Context) {
	var req service.UpdatePaymentConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err := h.configService.UpdatePaymentConfig(c.Request.Context(), req); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "updated"})
}
