package routes

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSupportTicketRefundRouteContract(t *testing.T) {
	t.Parallel()
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current test file")
	}
	routesDir := filepath.Dir(currentFile)
	paymentContent, err := os.ReadFile(filepath.Join(routesDir, "payment.go"))
	if err != nil {
		t.Fatalf("read payment routes: %v", err)
	}
	userContent, err := os.ReadFile(filepath.Join(routesDir, "user.go"))
	if err != nil {
		t.Fatalf("read user routes: %v", err)
	}
	adminContent, err := os.ReadFile(filepath.Join(routesDir, "admin.go"))
	if err != nil {
		t.Fatalf("read admin routes: %v", err)
	}
	source := string(paymentContent) + string(userContent) + string(adminContent)
	for _, route := range []string{
		`orders.GET("/:id/refund-quote", paymentHandler.GetRefundQuote)`,
		`adminOrders.POST("/:id/refund/query", adminPaymentHandler.QueryAndFinalizeRefund)`,
		`tickets.POST("", h.SupportTicket.Create)`,
		`tickets.POST("/:id/cancel", h.SupportTicket.Cancel)`,
		`tickets.POST("/:id/refund/review", h.Admin.SupportTicket.ReviewRefund)`,
	} {
		if !strings.Contains(source, route) {
			t.Errorf("payment route contract missing %s", route)
		}
	}
	for _, removed := range []string{`/refund-tickets`, `paymentHandler.CreateRefund`, `adminPaymentHandler.ProcessRefund`} {
		if strings.Contains(string(paymentContent), removed) {
			t.Errorf("legacy direct refund route remains: %s", removed)
		}
	}
}
