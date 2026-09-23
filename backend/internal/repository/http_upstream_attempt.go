package repository

import (
	"context"
	"github.com/AsukaCC/EasySub2api/internal/pkg/servertiming"
	"io"
	"net/http"
	"sync"
)

// doUpstreamRequest owns cancellation for one attempt, without cancelling the
// caller's context (which may be detached for billing or reused for retries).
func doUpstreamRequest(client *http.Client, req *http.Request) (*http.Response, error) {
	ctx, cancel := context.WithCancel(req.Context())
	resp, err := servertiming.Do(client, req.WithContext(ctx))
	if err != nil {
		cancel()
		return resp, err
	}
	decompressResponseBody(resp)
	resp.Body = &cancelOnCloseBody{ReadCloser: resp.Body, cancel: cancel}
	return resp, nil
}

type cancelOnCloseBody struct {
	io.ReadCloser
	cancel context.CancelFunc
	once   sync.Once
	err    error
	readMu sync.Mutex
	closed bool
}

func (b *cancelOnCloseBody) Read(p []byte) (int, error) {
	b.readMu.Lock()
	defer b.readMu.Unlock()
	if b.closed {
		return 0, http.ErrBodyReadAfterClose
	}
	return b.ReadCloser.Read(p)
}

func (b *cancelOnCloseBody) Close() error {
	b.once.Do(func() {
		// Cancel before closing, including before closing a decompressor. In Go
		// 1.27 an early HTTP/1 close concurrent with Read can otherwise leave an
		// EOF waiter on a reused connection and stall subsequent responses.
		// Fully consumed responses have already released their transport request,
		// so cancelling here preserves normal keep-alive reuse.
		b.cancel()
		// Wait for an active read to observe cancellation before touching the
		// body or decompressor. Do not hold readMu while cancelling the request.
		b.readMu.Lock()
		defer b.readMu.Unlock()
		b.closed = true
		b.err = b.ReadCloser.Close()
	})
	return b.err
}
