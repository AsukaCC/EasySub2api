package routes

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestPaymentRefundRouteContract(t *testing.T) {
	t.Parallel()
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current test file")
	}
	content, err := os.ReadFile(filepath.Join(filepath.Dir(currentFile), "payment.go"))
	if err != nil {
		t.Fatalf("read payment routes: %v", err)
	}
	source := string(content)
	for _, route := range []string{
		`orders.GET("/:id/refund-quote", paymentHandler.GetRefundQuote)`,
		`orders.POST("/:id/refunds", paymentHandler.CreateRefund)`,
		`orders.POST("/:id/refund-tickets", paymentHandler.CreateRefundTicket)`,
		`authenticated.GET("/refund-tickets", paymentHandler.ListRefundTickets)`,
		`authenticated.POST("/refund-tickets/:id/cancel", paymentHandler.CancelRefundTicket)`,
		`adminGroup.GET("/refund-tickets", adminPaymentHandler.ListRefundTickets)`,
		`adminGroup.POST("/refund-tickets/:id/review", adminPaymentHandler.ReviewRefundTicket)`,
		`adminOrders.POST("/:id/refund/query", adminPaymentHandler.QueryAndFinalizeRefund)`,
	} {
		if !strings.Contains(source, route) {
			t.Errorf("payment route contract missing %s", route)
		}
	}
}
