package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/config"
	coderws "github.com/coder/websocket"
	"github.com/stretchr/testify/require"
)

func TestSelectiveWSContextPoolCapacity(t *testing.T) {
	cfg := &config.Config{}
	cfg.Gateway.OpenAIWS.ModeRouterV2Enabled = true
	cfg.Gateway.OpenAIWS.DynamicMaxConnsByAccountConcurrencyEnabled = true
	cfg.Gateway.OpenAIWS.MaxConnsPerAccount = 30
	cfg.Gateway.OpenAIWS.OAuthMaxConnsFactor = 5
	cfg.Gateway.OpenAIWS.APIKeyMaxConnsFactor = 2.5
	pool := newOpenAIWSConnPool(cfg)
	defer pool.Close()
	for _, tc := range []struct {
		kind              string
		concurrency, want int
	}{
		{AccountTypeOAuth, 2, 10}, {AccountTypeOAuth, 10, 30},
		{AccountTypeOAuth, 0, 0}, {AccountTypeOAuth, -1, 0},
		{AccountTypeAPIKey, 3, 8},
	} {
		require.Equal(t, tc.want, pool.effectiveMaxConnsByAccount(&Account{Type: tc.kind, Concurrency: tc.concurrency}))
	}
}

func TestSelectiveWSWaiterReselectsReleasedConnection(t *testing.T) {
	cfg := &config.Config{}
	cfg.Gateway.OpenAIWS.MaxConnsPerAccount = 2
	cfg.Gateway.OpenAIWS.MaxIdlePerAccount = 2
	cfg.Gateway.OpenAIWS.PoolTargetUtilization = 1
	pool := newOpenAIWSConnPool(cfg)
	defer pool.Close()
	account := &Account{ID: "account", Type: AccountTypeAPIKey, Platform: PlatformOpenAI}
	ap := pool.getOrCreateAccountPool(account.ID)
	first := newOpenAIWSConn("first", account.ID, nil, nil)
	second := newOpenAIWSConn("second", account.ID, nil, nil)
	require.True(t, first.tryAcquire())
	require.True(t, second.tryAcquire())
	ap.mu.Lock()
	ap.conns[first.id], ap.conns[second.id] = first, second
	ap.mu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	done := make(chan *openAIWSConnLease, 1)
	errs := make(chan error, 1)
	go func() {
		lease, err := pool.Acquire(ctx, openAIWSAcquireRequest{Account: account, WSURL: "wss://unused.invalid/responses"})
		done <- lease
		errs <- err
	}()
	require.Eventually(t, func() bool { return first.waiters.Load()+second.waiters.Load() == 1 }, time.Second, time.Millisecond)
	released := second
	if second.waiters.Load() > 0 {
		released = first
	}
	(&openAIWSConnLease{pool: pool, accountID: account.ID, conn: released}).Release()
	select {
	case lease := <-done:
		require.NoError(t, <-errs)
		require.NotNil(t, lease)
		require.Same(t, released, lease.conn)
		lease.Release()
	case <-ctx.Done():
		t.Fatal("waiter remained attached to the busy connection")
	}
	require.Zero(t, first.waiters.Load()+second.waiters.Load())
}

func TestSelectiveWSReaderAnswersIdlePing(t *testing.T) {
	ping := make(chan struct{})
	result := make(chan error, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := coderws.Accept(w, r, nil)
		if err != nil {
			result <- err
			return
		}
		defer conn.CloseNow()
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		go func() { _, _, _ = conn.Read(ctx) }()
		<-ping
		result <- conn.Ping(ctx)
	}))
	defer server.Close()
	ws, _, _, err := newDefaultOpenAIWSClientDialer().Dial(context.Background(), "ws"+strings.TrimPrefix(server.URL, "http"), nil, "")
	require.NoError(t, err)
	conn := newOpenAIWSConn("idle-ping", "account", ws, nil)
	defer conn.abort()
	close(ping)
	require.NoError(t, <-result)
	require.EqualValues(t, 1, conn.upstreamPingCount())
}

type selectivePreemptCloser struct {
	ctx    context.Context
	sent   chan error
	finish chan struct{}
}

func (c *selectivePreemptCloser) Close(code coderws.StatusCode, reason string) error {
	if code != coderws.StatusTryAgainLater || reason == "" {
		panic("invalid preemption close")
	}
	c.sent <- c.ctx.Err()
	<-c.finish
	return nil
}

func TestSelectiveWSPreemptionOwnershipAndCloseOrder(t *testing.T) {
	svc := &OpenAIGatewayService{}
	account := &Account{ID: "account", Platform: PlatformOpenAI, Type: AccountTypeOAuth}
	ctx := newOpenAIWSExecutionScopeTestContext(map[string]string{"thread-id": "thread"})
	groupID := "group"
	ctx.Set("api_key", &APIKey{ID: "key", GroupID: &groupID})
	closer := &selectivePreemptCloser{sent: make(chan error, 1), finish: make(chan struct{})}
	first, cleanupFirst, armed := svc.BeginOpenAIWSIngressSessionPreemptionWithClient(context.Background(), ctx, account, nil, closer)
	require.True(t, armed)
	defer cleanupFirst()
	closer.ctx = first
	second, cleanupSecond, armed := svc.BeginOpenAIWSIngressSessionPreemptionWithClient(context.Background(), ctx, account, nil, nil)
	require.True(t, armed)
	defer cleanupSecond()
	select {
	case err := <-closer.sent:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("missing close frame")
	}
	require.True(t, isOpenAIWSSessionPreempted(first))
	require.NoError(t, first.Err())
	close(closer.finish)
	require.Eventually(t, func() bool { return first.Err() != nil }, time.Second, time.Millisecond)
	cleanupFirst()
	_, cleanupThird, _ := svc.BeginOpenAIWSIngressSessionPreemptionWithClient(context.Background(), ctx, account, nil, nil)
	defer cleanupThird()
	require.ErrorIs(t, context.Cause(second), errOpenAIWSSessionPreempted, "old cleanup must not remove the current owner")
}

func TestSelectiveWSPreemptionScopesAndRetries(t *testing.T) {
	svc := &OpenAIGatewayService{}
	account := &Account{ID: "account", Platform: PlatformOpenAI, Type: AccountTypeOAuth}
	ctx := newOpenAIWSExecutionScopeTestContext(map[string]string{"thread-id": "thread"})
	groupID := "group"
	ctx.Set("api_key", &APIKey{ID: "key", GroupID: &groupID})
	main, cleanup, _ := svc.BeginOpenAIWSIngressSessionPreemption(context.Background(), ctx, account, nil)
	defer cleanup()
	nested, nestedCleanup, _ := svc.BeginOpenAIWSIngressSessionPreemption(main, ctx, account, nil)
	nestedCleanup()
	require.Same(t, main, nested)
	require.NoError(t, main.Err())
	ctx.Request.Header.Set(openAIWSTurnMetadataHeader, `{"thread_id":"thread","request_kind":"memory"}`)
	memory, memoryCleanup, _ := svc.BeginOpenAIWSIngressSessionPreemption(context.Background(), ctx, account, nil)
	defer memoryCleanup()
	require.NoError(t, main.Err())
	ctx.Set("api_key", &APIKey{ID: "other-key", GroupID: &groupID})
	_, otherCleanup, _ := svc.BeginOpenAIWSIngressSessionPreemption(context.Background(), ctx, account, nil)
	defer otherCleanup()
	require.NoError(t, memory.Err())
	require.NotEqual(t, openAIWSExecutionScopeSeed("key", "thread", "t|kind=memory", ""), openAIWSExecutionScopeSeed("key", "thread", "t", "kind=memory"))
}
