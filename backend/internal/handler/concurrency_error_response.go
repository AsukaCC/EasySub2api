package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

const statusClientClosedRequest = 499

// 网关准入拒绝的稳定错误码：让客户端/看板能区分“等待队列已满”与
// “并发槽位超限”两种同为 429 rate_limit_error 的拒绝原因。
const (
	gatewayQueueFullCode        = "gateway_queue_full"
	gatewayConcurrencyLimitCode = "gateway_concurrency_limit"
)

// concurrencyErrorResponse 返回 (status, errType, code, message)；code 为空表示
// 无专门的准入拒绝码，沿用 errType 的默认映射。
func concurrencyErrorResponse(err error, slotType string) (int, string, string, string) {
	var waitQueueFullErr *WaitQueueFullError
	if errors.As(err, &waitQueueFullErr) {
		return http.StatusTooManyRequests, "rate_limit_error", gatewayQueueFullCode,
			"Too many pending requests, please retry later"
	}

	var concurrencyErr *ConcurrencyError
	if errors.As(err, &concurrencyErr) {
		if concurrencyErr.SlotType != "" {
			slotType = concurrencyErr.SlotType
		}
		return http.StatusTooManyRequests, "rate_limit_error", gatewayConcurrencyLimitCode,
			fmt.Sprintf("Concurrency limit exceeded for %s, please retry later", slotType)
	}

	if errors.Is(err, context.Canceled) {
		return statusClientClosedRequest, "api_error", "", "context canceled"
	}

	return http.StatusServiceUnavailable, "api_error", "", "Service temporarily unavailable, please retry later"
}
