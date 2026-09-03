package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	dbent "github.com/AsukaCC/EasySub2api/ent"
	"github.com/AsukaCC/EasySub2api/ent/apikey"
	"github.com/AsukaCC/EasySub2api/ent/group"
	"github.com/AsukaCC/EasySub2api/ent/paymentorder"
	"github.com/AsukaCC/EasySub2api/ent/paymentrefund"
	"github.com/AsukaCC/EasySub2api/ent/supportticket"
	"github.com/AsukaCC/EasySub2api/ent/supportticketmessage"
	"github.com/AsukaCC/EasySub2api/ent/supportticketread"
	"github.com/AsukaCC/EasySub2api/ent/user"
	infraerrors "github.com/AsukaCC/EasySub2api/internal/pkg/errors"
)

const (
	SupportTicketCategoryAccount = "ACCOUNT"
	SupportTicketCategoryRefund  = "REFUND"

	SupportTicketStatusPendingAdmin = "PENDING_ADMIN"
	SupportTicketStatusPendingUser  = "PENDING_USER"
	SupportTicketStatusInProgress   = "IN_PROGRESS"
	SupportTicketStatusResolved     = "RESOLVED"
	SupportTicketStatusClosed       = "CLOSED"
	SupportTicketStatusCancelled    = "CANCELLED"

	SupportTicketRefundNone     = "NONE"
	SupportTicketRefundPending  = "PENDING"
	SupportTicketRefundApproved = "APPROVED"
	SupportTicketRefundRejected = "REJECTED"

	SupportTicketOriginUser  = "USER"
	SupportTicketOriginAdmin = "ADMIN"

	SupportTicketRoleUser   = "USER"
	SupportTicketRoleAdmin  = "ADMIN"
	SupportTicketRoleSystem = "SYSTEM"

	SupportTicketMessageComment = "COMMENT"
	SupportTicketMessageSystem  = "SYSTEM"

	supportTicketTitleMax   = 120
	supportTicketMessageMax = 4000
	supportTicketReopenTTL  = 7 * 24 * time.Hour
)

type SupportTicketService struct {
	client         *dbent.Client
	paymentService *PaymentService
	settingService *SettingService
}

func NewSupportTicketService(client *dbent.Client, paymentService *PaymentService, settingService *SettingService) *SupportTicketService {
	return &SupportTicketService{client: client, paymentService: paymentService, settingService: settingService}
}

type CreateSupportTicketInput struct {
	UserID   string
	Category string
	Title    string
	Message  string
	APIKeyID string
	OrderID  string
}

type SupportTicketListFilters struct {
	Category string
	Status   string
	Search   string
	Unread   bool
}

type SupportTicketListItem struct {
	*dbent.SupportTicket
	Unread   bool   `json:"unread"`
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`
}

type SupportTicketPage struct {
	Items []SupportTicketListItem `json:"items"`
	Total int                     `json:"total"`
}

type SupportTicketUserSummary struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

type SupportTicketPermissions struct {
	CanReply        bool `json:"can_reply"`
	CanCancel       bool `json:"can_cancel"`
	CanClose        bool `json:"can_close"`
	CanReopen       bool `json:"can_reopen"`
	CanReviewRefund bool `json:"can_review_refund"`
	CanRetryRefund  bool `json:"can_retry_refund"`
}

type SupportTicketDetail struct {
	Ticket      *dbent.SupportTicket          `json:"ticket"`
	Messages    []*dbent.SupportTicketMessage `json:"messages"`
	User        *SupportTicketUserSummary     `json:"user,omitempty"`
	Refund      *PaymentRefundResponse        `json:"refund,omitempty"`
	Permissions SupportTicketPermissions      `json:"permissions"`
}

type SupportTicketSummary struct {
	Total            int  `json:"total"`
	Unread           int  `json:"unread"`
	CanCreateAccount bool `json:"can_create_account"`
	CanCreateRefund  bool `json:"can_create_refund"`
	FeatureEnabled   bool `json:"feature_enabled"`
}

type ReviewSupportTicketRefundInput struct {
	TicketID   string
	ReviewerID string
	Decision   string
	Amount     *float64
	Message    string
}

type CreateAdminRefundTicketInput struct {
	OrderID    string
	ReviewerID string
	Amount     *float64
	Message    string
}

type SupportTicketRefundReviewResult struct {
	Ticket *dbent.SupportTicket   `json:"ticket"`
	Refund *PaymentRefundResponse `json:"refund,omitempty"`
}

func (s *SupportTicketService) creationFlags(ctx context.Context, userFacing bool) (bool, bool, bool) {
	if s.settingService == nil {
		return false, false, false
	}
	enabled := s.settingService.IsSupportTicketsEnabled(ctx)
	if userFacing {
		enabled = s.settingService.IsSupportTicketsUserAvailable(ctx)
	}
	return enabled,
		enabled,
		enabled && s.settingService.IsPaymentUserAvailable(ctx)
}

func (s *SupportTicketService) Create(ctx context.Context, input CreateSupportTicketInput) (*SupportTicketDetail, error) {
	category := strings.ToUpper(strings.TrimSpace(input.Category))
	message := strings.TrimSpace(input.Message)
	if message == "" || len([]rune(message)) > supportTicketMessageMax {
		return nil, infraerrors.BadRequest("INVALID_TICKET_MESSAGE", "ticket message is required and must be at most 4000 characters")
	}
	_, accountEnabled, refundEnabled := s.creationFlags(ctx, true)
	switch category {
	case SupportTicketCategoryAccount:
		if !accountEnabled {
			return nil, infraerrors.Forbidden("SUPPORT_TICKET_ACCOUNT_DISABLED", "account support tickets are disabled")
		}
		return s.createAccountTicket(ctx, input, message)
	case SupportTicketCategoryRefund:
		if !refundEnabled {
			return nil, infraerrors.Forbidden("SUPPORT_TICKET_REFUND_DISABLED", "refund support tickets are disabled")
		}
		return s.createRefundTicket(ctx, input, message)
	default:
		return nil, infraerrors.BadRequest("INVALID_TICKET_CATEGORY", "category must be ACCOUNT or REFUND")
	}
}

func (s *SupportTicketService) CreateAdminRefund(ctx context.Context, input CreateAdminRefundTicketInput) (*SupportTicketRefundReviewResult, error) {
	_, _, refundEnabled := s.creationFlags(ctx, false)
	if !refundEnabled {
		return nil, infraerrors.Forbidden("SUPPORT_TICKET_REFUND_DISABLED", "refund support tickets are disabled")
	}
	orderID := strings.TrimSpace(input.OrderID)
	message := strings.TrimSpace(input.Message)
	if orderID == "" {
		return nil, infraerrors.BadRequest("ORDER_REQUIRED", "order_id is required")
	}
	if message == "" || len([]rune(message)) > supportTicketMessageMax {
		return nil, infraerrors.BadRequest("INVALID_TICKET_MESSAGE", "refund reason is required and must be at most 4000 characters")
	}
	order, err := s.client.PaymentOrder.Query().Where(paymentorder.IDEQ(orderID)).Only(ctx)
	if err != nil {
		return nil, infraerrors.NotFound("ORDER_NOT_FOUND", "order not found")
	}
	if _, err = s.paymentService.GetRefundQuote(ctx, orderID, order.UserID, input.Amount); err != nil {
		return nil, err
	}
	ticket, err := s.createTicketWithMessage(ctx, createTicketRecord{
		UserID: order.UserID, Category: SupportTicketCategoryRefund, Status: SupportTicketStatusPendingAdmin,
		Origin: SupportTicketOriginAdmin, Title: "Refund request for order #" + orderID,
		OrderID: orderID, RefundDecision: SupportTicketRefundPending,
		AuthorID: input.ReviewerID, AuthorRole: SupportTicketRoleAdmin,
	}, message)
	if err != nil {
		if dbent.IsConstraintError(err) {
			return nil, infraerrors.Conflict("ACTIVE_REFUND_TICKET_EXISTS", "an active refund ticket already exists for this order")
		}
		return nil, err
	}
	return s.ReviewRefund(ctx, ReviewSupportTicketRefundInput{
		TicketID: ticket.ID, ReviewerID: input.ReviewerID, Decision: "APPROVE", Amount: input.Amount, Message: message,
	})
}

func (s *SupportTicketService) createAccountTicket(ctx context.Context, input CreateSupportTicketInput, message string) (*SupportTicketDetail, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" || len([]rune(title)) > supportTicketTitleMax {
		return nil, infraerrors.BadRequest("INVALID_TICKET_TITLE", "title is required and must be at most 120 characters")
	}
	var keyID, keyName, groupID, groupName string
	if keyID = strings.TrimSpace(input.APIKeyID); keyID != "" {
		key, err := s.client.APIKey.Query().Where(apikey.IDEQ(keyID), apikey.UserIDEQ(input.UserID)).Only(ctx)
		if err != nil {
			return nil, infraerrors.NotFound("API_KEY_NOT_FOUND", "API key not found")
		}
		keyName = key.Name
		if key.GroupID != nil {
			groupID = *key.GroupID
			if g, groupErr := s.client.Group.Query().Where(group.IDEQ(groupID)).Only(ctx); groupErr == nil {
				groupName = g.Name
			}
		}
	}
	ticket, err := s.createTicketWithMessage(ctx, createTicketRecord{
		UserID: input.UserID, Category: SupportTicketCategoryAccount, Status: SupportTicketStatusPendingAdmin,
		Origin: SupportTicketOriginUser, Title: title, APIKeyID: keyID, APIKeyName: keyName,
		GroupID: groupID, GroupName: groupName, RefundDecision: SupportTicketRefundNone,
		AuthorID: input.UserID, AuthorRole: SupportTicketRoleUser,
	}, message)
	if err != nil {
		return nil, err
	}
	return s.GetUserDetail(ctx, ticket.ID, input.UserID)
}

func (s *SupportTicketService) createRefundTicket(ctx context.Context, input CreateSupportTicketInput, message string) (*SupportTicketDetail, error) {
	orderID := strings.TrimSpace(input.OrderID)
	if orderID == "" {
		return nil, infraerrors.BadRequest("ORDER_REQUIRED", "order_id is required for refund tickets")
	}
	quote, err := s.paymentService.GetRefundQuote(ctx, orderID, input.UserID, nil)
	if err != nil {
		return nil, err
	}
	ticket, err := s.createTicketWithMessage(ctx, createTicketRecord{
		UserID: input.UserID, Category: SupportTicketCategoryRefund, Status: SupportTicketStatusPendingAdmin,
		Origin: SupportTicketOriginUser, Title: "Refund request for order #" + orderID,
		OrderID: orderID, RefundDecision: SupportTicketRefundPending,
		AuthorID: input.UserID, AuthorRole: SupportTicketRoleUser,
	}, message)
	if err != nil {
		if dbent.IsConstraintError(err) {
			return nil, infraerrors.Conflict("ACTIVE_REFUND_TICKET_EXISTS", "an active refund ticket already exists for this order")
		}
		return nil, err
	}
	if quote.SelfServiceEligible {
		refund, refundErr := s.paymentService.CreateSelfServiceRefund(ctx, CreatePaymentRefundInput{
			OrderID: orderID, UserID: input.UserID, TicketID: ticket.ID,
			IdempotencyKey: "support-ticket-auto:" + ticket.ID, Reason: message,
		})
		if refundErr != nil {
			_ = s.addSystemEvent(ctx, ticket.ID, "REFUND_AUTO_FAILED", map[string]any{"message": refundErr.Error()}, true)
		} else {
			_ = s.syncRefundTicket(ctx, ticket.ID, refund)
		}
	}
	return s.GetUserDetail(ctx, ticket.ID, input.UserID)
}

type createTicketRecord struct {
	UserID, Category, Status, Origin, Title  string
	APIKeyID, APIKeyName, GroupID, GroupName string
	OrderID, RefundDecision                  string
	AuthorID, AuthorRole                     string
}

func (s *SupportTicketService) createTicketWithMessage(ctx context.Context, record createTicketRecord, body string) (_ *dbent.SupportTicket, err error) {
	tx, err := s.client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin support ticket: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	txCtx := dbent.NewTxContext(ctx, tx)
	now := time.Now()
	builder := tx.SupportTicket.Create().SetUserID(record.UserID).SetCategory(record.Category).
		SetStatus(record.Status).SetOrigin(record.Origin).SetTitle(record.Title).
		SetRefundDecision(record.RefundDecision)
	if record.AuthorRole == SupportTicketRoleAdmin {
		builder.SetLastAdminActivityAt(now)
	} else {
		builder.SetLastUserActivityAt(now)
	}
	if record.APIKeyID != "" {
		builder.SetAPIKeyID(record.APIKeyID)
	}
	if record.APIKeyName != "" {
		builder.SetAPIKeyNameSnapshot(record.APIKeyName)
	}
	if record.GroupID != "" {
		builder.SetGroupID(record.GroupID)
	}
	if record.GroupName != "" {
		builder.SetGroupNameSnapshot(record.GroupName)
	}
	if record.OrderID != "" {
		builder.SetOrderID(record.OrderID)
	}
	ticket, err := builder.Save(txCtx)
	if err != nil {
		return nil, fmt.Errorf("create support ticket: %w", err)
	}
	if _, err = tx.SupportTicketMessage.Create().SetTicketID(ticket.ID).SetAuthorID(record.AuthorID).
		SetAuthorRole(record.AuthorRole).SetKind(SupportTicketMessageComment).SetBody(body).Save(txCtx); err != nil {
		return nil, fmt.Errorf("create support ticket message: %w", err)
	}
	if _, err = tx.SupportTicketRead.Create().SetTicketID(ticket.ID).SetReaderID(record.AuthorID).SetReadAt(now).Save(txCtx); err != nil {
		return nil, fmt.Errorf("create support ticket read state: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit support ticket: %w", err)
	}
	return ticket, nil
}

func normalizeSupportTicketPage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func normalizeTicketFilter(value string, allowed ...string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	for _, candidate := range allowed {
		if value == candidate {
			return value
		}
	}
	return ""
}

func (s *SupportTicketService) ListUser(ctx context.Context, userID string, filters SupportTicketListFilters, page, pageSize int) (*SupportTicketPage, error) {
	page, pageSize = normalizeSupportTicketPage(page, pageSize)
	query := s.client.SupportTicket.Query().Where(supportticket.UserIDEQ(userID))
	applySupportTicketFilters(query, filters)
	return s.list(ctx, query, userID, page, pageSize, filters.Unread)
}

func (s *SupportTicketService) ListAdmin(ctx context.Context, adminID string, filters SupportTicketListFilters, page, pageSize int) (*SupportTicketPage, error) {
	page, pageSize = normalizeSupportTicketPage(page, pageSize)
	query := s.client.SupportTicket.Query()
	if search := strings.TrimSpace(filters.Search); search != "" {
		userIDs, _ := s.client.User.Query().Where(user.Or(user.EmailContainsFold(search), user.UsernameContainsFold(search))).IDs(ctx)
		orderIDs, _ := s.client.PaymentOrder.Query().Where(paymentorder.OutTradeNoContainsFold(search)).IDs(ctx)
		switch {
		case len(userIDs) > 0 && len(orderIDs) > 0:
			query.Where(supportticket.Or(supportticket.TitleContainsFold(search), supportticket.UserIDIn(userIDs...), supportticket.OrderIDIn(orderIDs...)))
		case len(userIDs) > 0:
			query.Where(supportticket.Or(supportticket.TitleContainsFold(search), supportticket.UserIDIn(userIDs...)))
		case len(orderIDs) > 0:
			query.Where(supportticket.Or(supportticket.TitleContainsFold(search), supportticket.OrderIDIn(orderIDs...)))
		default:
			query.Where(supportticket.TitleContainsFold(search))
		}
		filters.Search = ""
	}
	applySupportTicketFilters(query, filters)
	return s.list(ctx, query, adminID, page, pageSize, filters.Unread)
}

func applySupportTicketFilters(query *dbent.SupportTicketQuery, filters SupportTicketListFilters) {
	if category := normalizeTicketFilter(filters.Category, SupportTicketCategoryAccount, SupportTicketCategoryRefund); category != "" {
		query.Where(supportticket.CategoryEQ(category))
	}
	if status := normalizeTicketFilter(filters.Status, SupportTicketStatusPendingAdmin, SupportTicketStatusPendingUser, SupportTicketStatusInProgress, SupportTicketStatusResolved, SupportTicketStatusClosed, SupportTicketStatusCancelled); status != "" {
		query.Where(supportticket.StatusEQ(status))
	}
	if search := strings.TrimSpace(filters.Search); search != "" {
		query.Where(supportticket.TitleContainsFold(search))
	}
}

func (s *SupportTicketService) list(ctx context.Context, query *dbent.SupportTicketQuery, readerID string, page, pageSize int, unreadOnly bool) (*SupportTicketPage, error) {
	total := 0
	if !unreadOnly {
		var err error
		total, err = query.Clone().Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("count support tickets: %w", err)
		}
	}
	ordered := query.Order(supportticket.ByUpdatedAt(sql.OrderDesc()))
	if !unreadOnly {
		ordered.Offset((page - 1) * pageSize).Limit(pageSize)
	}
	items, err := ordered.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list support tickets: %w", err)
	}
	reads, err := s.readTimes(ctx, readerID, ticketIDs(items))
	if err != nil {
		return nil, err
	}
	result := make([]SupportTicketListItem, 0, len(items))
	for _, ticket := range items {
		unread := ticketUnread(ticket, reads[ticket.ID], readerID == ticket.UserID)
		if unreadOnly && !unread {
			continue
		}
		result = append(result, SupportTicketListItem{SupportTicket: ticket, Unread: unread})
	}
	if unreadOnly {
		total = len(result)
		start := (page - 1) * pageSize
		if start >= len(result) {
			result = []SupportTicketListItem{}
		} else {
			end := start + pageSize
			if end > len(result) {
				end = len(result)
			}
			result = result[start:end]
		}
	}
	s.attachListUsers(ctx, result)
	return &SupportTicketPage{Items: result, Total: total}, nil
}

func (s *SupportTicketService) attachListUsers(ctx context.Context, items []SupportTicketListItem) {
	ids := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		if item.SupportTicket == nil || item.UserID == "" {
			continue
		}
		if _, ok := seen[item.UserID]; ok {
			continue
		}
		seen[item.UserID] = struct{}{}
		ids = append(ids, item.UserID)
	}
	if len(ids) == 0 {
		return
	}
	users, err := s.client.User.Query().Where(user.IDIn(ids...)).All(ctx)
	if err != nil {
		return
	}
	byID := make(map[string]*dbent.User, len(users))
	for _, item := range users {
		byID[item.ID] = item
	}
	for i := range items {
		if items[i].SupportTicket == nil {
			continue
		}
		owner := byID[items[i].UserID]
		if owner == nil {
			continue
		}
		items[i].Username = owner.Username
		items[i].Email = owner.Email
	}
}

func ticketIDs(items []*dbent.SupportTicket) []string {
	ids := make([]string, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	return ids
}

func (s *SupportTicketService) readTimes(ctx context.Context, readerID string, ids []string) (map[string]time.Time, error) {
	result := map[string]time.Time{}
	if len(ids) == 0 {
		return result, nil
	}
	reads, err := s.client.SupportTicketRead.Query().Where(supportticketread.ReaderIDEQ(readerID), supportticketread.TicketIDIn(ids...)).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("load support ticket reads: %w", err)
	}
	for _, read := range reads {
		result[read.TicketID] = read.ReadAt
	}
	return result, nil
}

func ticketUnread(ticket *dbent.SupportTicket, readAt time.Time, readerIsOwner bool) bool {
	activity := ticket.LastUserActivityAt
	if readerIsOwner {
		activity = ticket.LastAdminActivityAt
	}
	return activity != nil && (readAt.IsZero() || activity.After(readAt))
}

func (s *SupportTicketService) Summary(ctx context.Context, readerID string, admin bool) (*SupportTicketSummary, error) {
	query := s.client.SupportTicket.Query()
	if !admin {
		query.Where(supportticket.UserIDEQ(readerID))
	}
	items, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("load support ticket summary: %w", err)
	}
	reads, err := s.readTimes(ctx, readerID, ticketIDs(items))
	if err != nil {
		return nil, err
	}
	unread := 0
	for _, ticket := range items {
		if ticketUnread(ticket, reads[ticket.ID], !admin) {
			unread++
		}
	}
	enabled, account, refund := s.creationFlags(ctx, !admin)
	return &SupportTicketSummary{Total: len(items), Unread: unread, FeatureEnabled: enabled, CanCreateAccount: account, CanCreateRefund: refund}, nil
}

func (s *SupportTicketService) GetUserDetail(ctx context.Context, ticketID, userID string) (*SupportTicketDetail, error) {
	ticket, err := s.client.SupportTicket.Query().Where(supportticket.IDEQ(ticketID), supportticket.UserIDEQ(userID)).Only(ctx)
	if err != nil {
		return nil, infraerrors.NotFound("SUPPORT_TICKET_NOT_FOUND", "support ticket not found")
	}
	return s.detail(ctx, ticket, false)
}

func (s *SupportTicketService) GetAdminDetail(ctx context.Context, ticketID string) (*SupportTicketDetail, error) {
	ticket, err := s.client.SupportTicket.Get(ctx, ticketID)
	if err != nil {
		return nil, infraerrors.NotFound("SUPPORT_TICKET_NOT_FOUND", "support ticket not found")
	}
	return s.detail(ctx, ticket, true)
}

func (s *SupportTicketService) detail(ctx context.Context, ticket *dbent.SupportTicket, admin bool) (*SupportTicketDetail, error) {
	messages, err := s.client.SupportTicketMessage.Query().Where(supportticketmessage.TicketIDEQ(ticket.ID)).Order(supportticketmessage.ByCreatedAt()).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("load support ticket messages: %w", err)
	}
	detail := &SupportTicketDetail{Ticket: ticket, Messages: messages, Permissions: supportTicketPermissions(ticket, admin)}
	if admin {
		if user, userErr := s.client.User.Get(ctx, ticket.UserID); userErr == nil {
			detail.User = &SupportTicketUserSummary{ID: user.ID, Email: user.Email, Username: user.Username}
		}
	}
	if ticket.RefundID != nil {
		if refund, refundErr := s.client.PaymentRefund.Get(ctx, *ticket.RefundID); refundErr == nil {
			detail.Refund = paymentRefundResponse(refund)
		}
	}
	return detail, nil
}

func supportTicketPermissions(ticket *dbent.SupportTicket, admin bool) SupportTicketPermissions {
	terminal := ticket.Status == SupportTicketStatusResolved || ticket.Status == SupportTicketStatusClosed || ticket.Status == SupportTicketStatusCancelled
	p := SupportTicketPermissions{CanReply: !terminal, CanReviewRefund: admin && ticket.Category == SupportTicketCategoryRefund && !terminal && ticket.RefundDecision == SupportTicketRefundPending}
	p.CanRetryRefund = admin && ticket.Category == SupportTicketCategoryRefund && ticket.Status == SupportTicketStatusPendingAdmin && ticket.RefundDecision == SupportTicketRefundApproved && ticket.ApprovedPrincipalAmount != nil
	if admin {
		return p
	}
	p.CanCancel = ticket.Status == SupportTicketStatusPendingAdmin && ticket.LastAdminActivityAt == nil
	p.CanClose = ticket.Status == SupportTicketStatusPendingUser || ticket.Status == SupportTicketStatusResolved
	p.CanReopen = ticket.Status == SupportTicketStatusResolved && ticket.ReopenCount == 0 && ticket.ResolvedAt != nil && time.Since(*ticket.ResolvedAt) <= supportTicketReopenTTL
	return p
}

func (s *SupportTicketService) MarkRead(ctx context.Context, ticketID, readerID string, admin bool) error {
	if !admin {
		if _, err := s.client.SupportTicket.Query().Where(supportticket.IDEQ(ticketID), supportticket.UserIDEQ(readerID)).Only(ctx); err != nil {
			return infraerrors.NotFound("SUPPORT_TICKET_NOT_FOUND", "support ticket not found")
		}
	} else if _, err := s.client.SupportTicket.Get(ctx, ticketID); err != nil {
		return infraerrors.NotFound("SUPPORT_TICKET_NOT_FOUND", "support ticket not found")
	}
	return s.client.SupportTicketRead.Create().SetTicketID(ticketID).SetReaderID(readerID).SetReadAt(time.Now()).
		OnConflictColumns(supportticketread.FieldTicketID, supportticketread.FieldReaderID).UpdateReadAt().Exec(ctx)
}

func (s *SupportTicketService) Reply(ctx context.Context, ticketID, authorID, role, body string) (*SupportTicketDetail, error) {
	body = strings.TrimSpace(body)
	if body == "" || len([]rune(body)) > supportTicketMessageMax {
		return nil, infraerrors.BadRequest("INVALID_TICKET_MESSAGE", "message is required and must be at most 4000 characters")
	}
	ticket, err := s.client.SupportTicket.Get(ctx, ticketID)
	if err != nil || (role == SupportTicketRoleUser && ticket.UserID != authorID) {
		return nil, infraerrors.NotFound("SUPPORT_TICKET_NOT_FOUND", "support ticket not found")
	}
	if ticket.Status == SupportTicketStatusClosed || ticket.Status == SupportTicketStatusCancelled || ticket.Status == SupportTicketStatusResolved {
		return nil, infraerrors.Conflict("SUPPORT_TICKET_STATE_CONFLICT", "ticket cannot receive replies in its current state")
	}
	now := time.Now()
	nextStatus := SupportTicketStatusPendingAdmin
	update := s.client.SupportTicket.UpdateOneID(ticketID)
	if role == SupportTicketRoleAdmin {
		nextStatus = SupportTicketStatusPendingUser
		update.SetLastAdminActivityAt(now)
	} else {
		update.SetLastUserActivityAt(now)
	}
	tx, err := s.client.Tx(ctx)
	if err != nil {
		return nil, err
	}
	txCtx := dbent.NewTxContext(ctx, tx)
	if _, err = tx.SupportTicketMessage.Create().SetTicketID(ticketID).SetAuthorID(authorID).SetAuthorRole(role).SetKind(SupportTicketMessageComment).SetBody(body).Save(txCtx); err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	txUpdate := tx.SupportTicket.UpdateOneID(ticketID).SetStatus(nextStatus)
	if role == SupportTicketRoleAdmin {
		txUpdate.SetLastAdminActivityAt(now)
	} else {
		txUpdate.SetLastUserActivityAt(now)
	}
	if _, err = txUpdate.Save(txCtx); err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	_ = update
	if role == SupportTicketRoleAdmin {
		return s.GetAdminDetail(ctx, ticketID)
	}
	return s.GetUserDetail(ctx, ticketID, authorID)
}

func (s *SupportTicketService) UserAction(ctx context.Context, ticketID, userID, action string) (*SupportTicketDetail, error) {
	ticket, err := s.client.SupportTicket.Query().Where(supportticket.IDEQ(ticketID), supportticket.UserIDEQ(userID)).Only(ctx)
	if err != nil {
		return nil, infraerrors.NotFound("SUPPORT_TICKET_NOT_FOUND", "support ticket not found")
	}
	now := time.Now()
	update := s.client.SupportTicket.UpdateOneID(ticket.ID)
	event := ""
	switch strings.ToUpper(action) {
	case "CANCEL":
		if ticket.Status != SupportTicketStatusPendingAdmin || ticket.LastAdminActivityAt != nil {
			return nil, infraerrors.Conflict("SUPPORT_TICKET_STATE_CONFLICT", "only untouched pending tickets can be cancelled")
		}
		update.SetStatus(SupportTicketStatusCancelled).SetClosedAt(now).SetLastUserActivityAt(now)
		event = "TICKET_CANCELLED"
	case "CLOSE":
		if ticket.Status != SupportTicketStatusPendingUser && ticket.Status != SupportTicketStatusResolved {
			return nil, infraerrors.Conflict("SUPPORT_TICKET_STATE_CONFLICT", "ticket cannot be closed")
		}
		update.SetStatus(SupportTicketStatusClosed).SetClosedAt(now).SetLastUserActivityAt(now)
		event = "TICKET_CLOSED"
	case "REOPEN":
		if ticket.Status != SupportTicketStatusResolved || ticket.ReopenCount != 0 || ticket.ResolvedAt == nil || now.Sub(*ticket.ResolvedAt) > supportTicketReopenTTL {
			return nil, infraerrors.Conflict("SUPPORT_TICKET_REOPEN_EXPIRED", "ticket can only be reopened once within 7 days")
		}
		update.SetStatus(SupportTicketStatusPendingAdmin).SetReopenCount(1).ClearResolvedAt().SetLastUserActivityAt(now)
		event = "TICKET_REOPENED"
	default:
		return nil, infraerrors.BadRequest("INVALID_TICKET_ACTION", "invalid ticket action")
	}
	if _, err := update.Save(ctx); err != nil {
		return nil, err
	}
	_ = s.addSystemEvent(ctx, ticket.ID, event, nil, false)
	return s.GetUserDetail(ctx, ticket.ID, userID)
}

func (s *SupportTicketService) AdminSetStatus(ctx context.Context, ticketID, adminID, status, message string) (*SupportTicketDetail, error) {
	status = normalizeTicketFilter(status, SupportTicketStatusPendingAdmin, SupportTicketStatusPendingUser, SupportTicketStatusInProgress, SupportTicketStatusResolved, SupportTicketStatusClosed)
	if status == "" {
		return nil, infraerrors.BadRequest("INVALID_TICKET_STATUS", "invalid ticket status")
	}
	ticket, err := s.client.SupportTicket.Get(ctx, ticketID)
	if err != nil {
		return nil, infraerrors.NotFound("SUPPORT_TICKET_NOT_FOUND", "support ticket not found")
	}
	if ticket.Status == SupportTicketStatusClosed || ticket.Status == SupportTicketStatusCancelled {
		return nil, infraerrors.Conflict("SUPPORT_TICKET_STATE_CONFLICT", "terminal ticket cannot change status")
	}
	if ticket.Category == SupportTicketCategoryRefund && s.paymentService != nil && (status == SupportTicketStatusResolved || status == SupportTicketStatusClosed) {
		refund, reconcileErr := s.paymentService.ReconcileRefundForTicket(ctx, ticket.ID, psStringValue(ticket.OrderID))
		if reconcileErr != nil {
			return nil, reconcileErr
		}
		if refund != nil && refund.Status != RefundStatusSucceeded && refund.Status != RefundStatusFailed {
			return nil, infraerrors.Conflict("REFUND_STILL_PENDING", "the payment provider still reports this refund as pending")
		}
	}
	now := time.Now()
	update := s.client.SupportTicket.UpdateOneID(ticketID).SetStatus(status).SetLastAdminActivityAt(now)
	if status == SupportTicketStatusResolved {
		update.SetResolvedAt(now)
	}
	if status == SupportTicketStatusClosed {
		update.SetClosedAt(now)
	}
	if _, err = update.Save(ctx); err != nil {
		return nil, err
	}
	if body := strings.TrimSpace(message); body != "" {
		if _, err = s.client.SupportTicketMessage.Create().SetTicketID(ticketID).SetAuthorID(adminID).
			SetAuthorRole(SupportTicketRoleAdmin).SetKind(SupportTicketMessageComment).SetBody(body).Save(ctx); err != nil {
			return nil, err
		}
	}
	_ = s.addSystemEvent(ctx, ticketID, "STATUS_CHANGED", map[string]any{"status": status}, true)
	return s.GetAdminDetail(ctx, ticketID)
}

func (s *SupportTicketService) ReviewRefund(ctx context.Context, input ReviewSupportTicketRefundInput) (*SupportTicketRefundReviewResult, error) {
	ticket, err := s.client.SupportTicket.Get(ctx, input.TicketID)
	if err != nil {
		return nil, infraerrors.NotFound("SUPPORT_TICKET_NOT_FOUND", "support ticket not found")
	}
	if ticket.Category != SupportTicketCategoryRefund || ticket.OrderID == nil {
		return nil, infraerrors.BadRequest("INVALID_TICKET_CATEGORY", "ticket is not a refund ticket")
	}
	decision := strings.ToUpper(strings.TrimSpace(input.Decision))
	if decision == "RETRY" {
		if ticket.Status != SupportTicketStatusPendingAdmin || ticket.ApprovedPrincipalAmount == nil {
			return nil, infraerrors.BadRequest("REFUND_RETRY_NOT_ALLOWED", "only a failed approved refund can be retried")
		}
		input.Amount = ticket.ApprovedPrincipalAmount
		decision = "APPROVE"
	}
	now := time.Now()
	if decision == "REJECT" {
		_, err = s.client.SupportTicket.UpdateOneID(ticket.ID).SetRefundDecision(SupportTicketRefundRejected).SetStatus(SupportTicketStatusResolved).
			SetReviewerID(input.ReviewerID).SetReviewedAt(now).SetResolvedAt(now).SetLastAdminActivityAt(now).Save(ctx)
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(input.Message) != "" {
			_, _ = s.client.SupportTicketMessage.Create().SetTicketID(ticket.ID).SetAuthorID(input.ReviewerID).SetAuthorRole(SupportTicketRoleAdmin).SetKind(SupportTicketMessageComment).SetBody(strings.TrimSpace(input.Message)).Save(ctx)
		}
		_ = s.addSystemEvent(ctx, ticket.ID, "REFUND_REJECTED", nil, true)
		updated, _ := s.client.SupportTicket.Get(ctx, ticket.ID)
		return &SupportTicketRefundReviewResult{Ticket: updated}, nil
	}
	if decision != "APPROVE" || input.Amount == nil || *input.Amount <= 0 || math.IsNaN(*input.Amount) || math.IsInf(*input.Amount, 0) || !hasRefundMoneyPrecision(*input.Amount) {
		return nil, infraerrors.BadRequest("INVALID_REFUND_DECISION", "approval requires a positive amount with at most 2 decimal places")
	}
	_, err = s.client.SupportTicket.UpdateOneID(ticket.ID).SetRefundDecision(SupportTicketRefundApproved).SetStatus(SupportTicketStatusInProgress).
		SetApprovedPrincipalAmount(*input.Amount).SetReviewerID(input.ReviewerID).SetReviewedAt(now).SetLastAdminActivityAt(now).Save(ctx)
	if err != nil {
		return nil, err
	}
	attempt, _ := s.client.PaymentRefund.Query().Where(paymentrefund.TicketIDEQ(ticket.ID)).Count(ctx)
	reason := strings.TrimSpace(input.Message)
	if reason == "" {
		reason = "approved support ticket refund"
	}
	refund, _, err := s.paymentService.preparePaymentRefund(ctx, CreatePaymentRefundInput{
		OrderID: *ticket.OrderID, UserID: ticket.UserID, RequestedBy: input.ReviewerID,
		IdempotencyKey: fmt.Sprintf("support-ticket:%s:%d", ticket.ID, attempt+1), Principal: input.Amount,
		Reason: reason, Source: RefundSourceTicket, TicketID: ticket.ID, AutoAffiliate: false,
	})
	if err != nil {
		_, _ = s.client.SupportTicket.UpdateOneID(ticket.ID).SetStatus(SupportTicketStatusPendingAdmin).Save(ctx)
		return nil, err
	}
	if refund.Status == RefundStatusReserved || refund.Status == RefundStatusSubmitting {
		refund, err = s.paymentService.executePaymentRefund(ctx, refund.ID)
	}
	response := paymentRefundResponse(refund)
	_ = s.syncRefundTicket(ctx, ticket.ID, response)
	if strings.TrimSpace(input.Message) != "" {
		_, _ = s.client.SupportTicketMessage.Create().SetTicketID(ticket.ID).SetAuthorID(input.ReviewerID).SetAuthorRole(SupportTicketRoleAdmin).SetKind(SupportTicketMessageComment).SetBody(strings.TrimSpace(input.Message)).Save(ctx)
	}
	_ = s.addSystemEvent(ctx, ticket.ID, "REFUND_APPROVED", map[string]any{"amount": *input.Amount}, true)
	updated, _ := s.client.SupportTicket.Get(ctx, ticket.ID)
	return &SupportTicketRefundReviewResult{Ticket: updated, Refund: response}, err
}

func (s *SupportTicketService) syncRefundTicket(ctx context.Context, ticketID string, refund *PaymentRefundResponse) error {
	if refund == nil {
		return nil
	}
	now := time.Now()
	update := s.client.SupportTicket.UpdateOneID(ticketID).SetRefundID(refund.ID).SetRefundDecision(SupportTicketRefundApproved).SetLastAdminActivityAt(now)
	event := "REFUND_PROCESSING"
	switch refund.Status {
	case RefundStatusSucceeded:
		update.SetStatus(SupportTicketStatusResolved).SetResolvedAt(now)
		event = "REFUND_COMPLETED"
	case RefundStatusFailed:
		update.SetStatus(SupportTicketStatusPendingAdmin)
		event = "REFUND_FAILED"
	default:
		update.SetStatus(SupportTicketStatusInProgress)
	}
	if _, err := update.Save(ctx); err != nil {
		return err
	}
	return s.addSystemEvent(ctx, ticketID, event, map[string]any{"refund_id": refund.ID, "status": refund.Status}, true)
}

func (s *SupportTicketService) addSystemEvent(ctx context.Context, ticketID, eventType string, data map[string]any, notifyUser bool) error {
	raw := "{}"
	if data != nil {
		if encoded, err := json.Marshal(data); err == nil {
			raw = string(encoded)
		}
	}
	if _, err := s.client.SupportTicketMessage.Create().SetTicketID(ticketID).SetAuthorRole(SupportTicketRoleSystem).SetKind(SupportTicketMessageSystem).SetEventType(eventType).SetEventData(raw).Save(ctx); err != nil {
		return err
	}
	if notifyUser {
		_, err := s.client.SupportTicket.UpdateOneID(ticketID).SetLastAdminActivityAt(time.Now()).Save(ctx)
		return err
	}
	return nil
}
