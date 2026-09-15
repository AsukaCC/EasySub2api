package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"go.uber.org/zap"

	"github.com/AsukaCC/EasySub2api/internal/pkg/ctxkey"
	"github.com/AsukaCC/EasySub2api/internal/pkg/logger"
	"github.com/AsukaCC/EasySub2api/internal/service"
)

const (
	upstreamHeaderTraceRequestIDEnv    = "EASYSUB2API_DEBUG_UPSTREAM_HEADERS_REQUEST_ID"
	upstreamHeaderTraceMaxAttemptsEnv  = "EASYSUB2API_DEBUG_UPSTREAM_HEADERS_MAX_ATTEMPTS"
	defaultUpstreamHeaderTraceAttempts = 16
	maxUpstreamHeaderTraceAttempts     = 64
	maxUpstreamHeaderTraceFields       = 128
	maxUpstreamHeaderTraceValueRunes   = 512
)

// upstreamHeaderTraceGate observes the logical header fields after the
// transport has normalized and supplemented them, including HTTP/2 pseudo
// headers. It is intentionally opt-in and scoped to one correlation ID.
type upstreamHeaderTraceGate struct {
	requestID   string
	maxAttempts uint32
	attempts    atomic.Uint32
}

type upstreamHeaderTraceCapture struct {
	ctx              context.Context
	requestID        string
	attempt          uint32
	accountID        string
	profile          service.HTTPUpstreamProfile
	protocolMode     string
	proxyKey         string
	method           string
	targetScheme     string
	targetHost       string
	targetPath       string
	requestHost      string
	contentLength    int64
	transferEncoding []string

	mu                    sync.Mutex
	headers               []string
	droppedHeaderValues   int
	headerWriteCompleted  bool
	requestWriteCompleted bool
	requestWriteError     bool
	finishOnce            sync.Once
}

type upstreamHeaderTraceSnapshot struct {
	headers               []string
	droppedHeaderValues   int
	headerWriteCompleted  bool
	requestWriteCompleted bool
	requestWriteError     bool
}

func newUpstreamHeaderTraceGateFromEnv() *upstreamHeaderTraceGate {
	requestID := strings.TrimSpace(os.Getenv(upstreamHeaderTraceRequestIDEnv))
	if requestID == "" {
		return nil
	}

	maxAttempts := defaultUpstreamHeaderTraceAttempts
	if raw := strings.TrimSpace(os.Getenv(upstreamHeaderTraceMaxAttemptsEnv)); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			maxAttempts = parsed
		}
	}
	if maxAttempts > maxUpstreamHeaderTraceAttempts {
		maxAttempts = maxUpstreamHeaderTraceAttempts
	}

	return &upstreamHeaderTraceGate{
		requestID:   requestID,
		maxAttempts: uint32(maxAttempts),
	}
}

func (s *httpUpstreamService) attachUpstreamHeaderTrace(
	req *http.Request,
	profile service.HTTPUpstreamProfile,
	accountID string,
	protocolMode string,
	proxyKey string,
) (*http.Request, *upstreamHeaderTraceCapture) {
	if s == nil || s.upstreamHeaderTrace == nil {
		return req, nil
	}
	return s.upstreamHeaderTrace.attach(req, profile, accountID, protocolMode, proxyKey)
}

func (g *upstreamHeaderTraceGate) attach(
	req *http.Request,
	profile service.HTTPUpstreamProfile,
	accountID string,
	protocolMode string,
	proxyKey string,
) (*http.Request, *upstreamHeaderTraceCapture) {
	if g == nil || req == nil {
		return req, nil
	}

	requestID, _ := req.Context().Value(ctxkey.RequestID).(string)
	if strings.TrimSpace(requestID) != g.requestID {
		return req, nil
	}

	attempt := g.attempts.Add(1)
	if attempt > g.maxAttempts {
		return req, nil
	}

	capture := newUpstreamHeaderTraceCapture(req, requestID, attempt, accountID, profile, protocolMode, proxyKey)
	trace := &httptrace.ClientTrace{
		WroteHeaderField: capture.recordHeaderField,
		WroteHeaders:     capture.markHeadersWritten,
		WroteRequest:     capture.markRequestWritten,
	}
	return req.WithContext(httptrace.WithClientTrace(req.Context(), trace)), capture
}

func newUpstreamHeaderTraceCapture(
	req *http.Request,
	requestID string,
	attempt uint32,
	accountID string,
	profile service.HTTPUpstreamProfile,
	protocolMode string,
	proxyKey string,
) *upstreamHeaderTraceCapture {
	capture := &upstreamHeaderTraceCapture{
		ctx:              req.Context(),
		requestID:        requestID,
		attempt:          attempt,
		accountID:        accountID,
		profile:          profile,
		protocolMode:     protocolMode,
		proxyKey:         proxyKey,
		method:           req.Method,
		contentLength:    req.ContentLength,
		transferEncoding: append([]string(nil), req.TransferEncoding...),
	}
	if req.URL != nil {
		capture.targetScheme = req.URL.Scheme
		capture.targetHost = req.URL.Host
		capture.targetPath = req.URL.EscapedPath()
		if capture.targetPath == "" {
			capture.targetPath = "/"
		}
	}
	capture.requestHost = req.Host
	if capture.requestHost == "" {
		capture.requestHost = capture.targetHost
	}
	return capture
}

func (c *upstreamHeaderTraceCapture) recordHeaderField(name string, values []string) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(values) == 0 {
		values = []string{""}
	}
	for _, value := range values {
		if len(c.headers) >= maxUpstreamHeaderTraceFields {
			c.droppedHeaderValues++
			continue
		}
		c.headers = append(c.headers, name+": "+sanitizeUpstreamHeaderTraceValue(name, value))
	}
}

func (c *upstreamHeaderTraceCapture) markHeadersWritten() {
	if c == nil {
		return
	}
	c.mu.Lock()
	c.headerWriteCompleted = true
	c.mu.Unlock()
}

func (c *upstreamHeaderTraceCapture) markRequestWritten(info httptrace.WroteRequestInfo) {
	if c == nil {
		return
	}
	c.mu.Lock()
	c.requestWriteCompleted = true
	c.requestWriteError = info.Err != nil
	c.mu.Unlock()
}

func (c *upstreamHeaderTraceCapture) snapshot() upstreamHeaderTraceSnapshot {
	if c == nil {
		return upstreamHeaderTraceSnapshot{}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return upstreamHeaderTraceSnapshot{
		headers:               append([]string(nil), c.headers...),
		droppedHeaderValues:   c.droppedHeaderValues,
		headerWriteCompleted:  c.headerWriteCompleted,
		requestWriteCompleted: c.requestWriteCompleted,
		requestWriteError:     c.requestWriteError,
	}
}

func (c *upstreamHeaderTraceCapture) finish(resp *http.Response, err error) {
	if c == nil {
		return
	}
	c.finishOnce.Do(func() {
		snapshot := c.snapshot()
		statusCode := 0
		responseProtocol := ""
		if resp != nil {
			statusCode = resp.StatusCode
			responseProtocol = resp.Proto
		}

		logger.FromContext(c.ctx).Info(
			"upstream_headers_written",
			zap.String("request_id", c.requestID),
			zap.Uint32("trace_attempt", c.attempt),
			zap.String("account_id_fingerprint", fingerprintUpstreamHeaderTraceValue(c.accountID)),
			zap.String("upstream_profile", string(c.profile)),
			zap.String("protocol_mode", c.protocolMode),
			zap.Bool("proxy_configured", c.proxyKey != "" && c.proxyKey != directProxyKey),
			zap.String("proxy_fingerprint", fingerprintUpstreamHeaderTraceValue(c.proxyKey)),
			zap.String("method", c.method),
			zap.String("target_scheme", c.targetScheme),
			zap.String("target_host", c.targetHost),
			zap.String("target_path", c.targetPath),
			zap.String("request_host", c.requestHost),
			zap.Int64("content_length", c.contentLength),
			zap.Strings("transfer_encoding", c.transferEncoding),
			zap.Bool("headers_written", snapshot.headerWriteCompleted),
			zap.Bool("request_write_completed", snapshot.requestWriteCompleted),
			zap.Bool("request_write_error", snapshot.requestWriteError),
			zap.Strings("written_headers", snapshot.headers),
			zap.Int("dropped_header_values", snapshot.droppedHeaderValues),
			zap.Int("response_status", statusCode),
			zap.String("response_protocol", responseProtocol),
			zap.Bool("transport_error", err != nil),
			zap.Bool(logger.OpsSystemLogSkipField, true),
		)
	})
}

func sanitizeUpstreamHeaderTraceValue(name string, value string) string {
	value = strings.ToValidUTF8(strings.TrimSpace(value), "?")
	lowerName := strings.ToLower(strings.TrimSpace(name))

	switch lowerName {
	case "authorization", "proxy-authorization", "x-api-key", "api-key", "x-goog-api-key":
		scheme := "credential"
		if fields := strings.Fields(value); len(fields) > 1 {
			scheme = fields[0]
		}
		return scheme + " " + summarizeSensitiveUpstreamHeaderTraceValue(value)
	case "cookie", "set-cookie", "chatgpt-account-id", "session_id", "conversation_id",
		"x-codex-installation-id", "x-codex-turn-state", "x-codex-turn-metadata", "x-codex-window-id":
		return summarizeSensitiveUpstreamHeaderTraceValue(value)
	case ":path":
		if parsed, err := url.ParseRequestURI(value); err == nil {
			value = parsed.EscapedPath()
		} else if path, _, found := strings.Cut(value, "?"); found {
			value = path
		}
		if value == "" {
			value = "/"
		}
	}

	if isSafeUpstreamHeaderTraceName(lowerName) {
		return truncateUpstreamHeaderTraceValue(value)
	}
	return summarizeSensitiveUpstreamHeaderTraceValue(value)
}

func isSafeUpstreamHeaderTraceName(name string) bool {
	switch name {
	case ":authority", ":method", ":path", ":scheme",
		"accept", "accept-encoding", "accept-language", "cache-control", "connection",
		"content-encoding", "content-length", "content-type", "host", "openai-beta",
		"originator", "te", "transfer-encoding", "user-agent", "version", "x-codex-beta-features",
		"x-codex-routing-hint", "x-openai-fedramp", "x-request-id":
		return true
	default:
		return false
	}
}

func summarizeSensitiveUpstreamHeaderTraceValue(value string) string {
	return "[redacted len=" + strconv.Itoa(len(value)) + " sha256=" + fingerprintUpstreamHeaderTraceValue(value) + "]"
}

func fingerprintUpstreamHeaderTraceValue(value string) string {
	if value == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:6])
}

func truncateUpstreamHeaderTraceValue(value string) string {
	runes := []rune(value)
	if len(runes) <= maxUpstreamHeaderTraceValueRunes {
		return value
	}
	return string(runes[:maxUpstreamHeaderTraceValueRunes]) + "..."
}
