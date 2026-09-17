package service

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/AsukaCC/EasySub2api/internal/pkg/tlsfingerprint"
	coderws "github.com/coder/websocket"
	"github.com/stretchr/testify/require"
)

func TestProtectionWebSocketActualTLSProfilesAndPoolKeys(t *testing.T) {
	hellos := make(chan []uint16, 4)
	target := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := coderws.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer conn.CloseNow()
		typ, body, err := conn.Read(r.Context())
		if err == nil {
			_ = conn.Write(r.Context(), typ, body)
		}
	}))
	target.TLS = &tls.Config{GetConfigForClient: func(info *tls.ClientHelloInfo) (*tls.Config, error) {
		hellos <- append([]uint16(nil), info.CipherSuites...)
		return nil, nil
	}}
	target.StartTLS()
	defer target.Close()
	roots := x509.NewCertPool()
	roots.AddCert(target.Certificate())
	dialer := &coderOpenAIWSClientDialer{tlsRootCAs: roots}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	url := strings.Replace(target.URL, "https:", "wss:", 1)
	for _, name := range []string{"nodejs24", "nodejs22"} {
		profile := tlsfingerprint.BuiltinProfile(name)
		conn, _, _, err := dialer.Dial(withOpenAIWSTLSProfile(ctx, "account", profile), url, http.Header{}, "")
		require.NoError(t, err)
		require.NoError(t, conn.WriteJSON(ctx, map[string]string{"type": "response.create"}))
		body, err := conn.ReadMessage(ctx)
		require.NoError(t, err)
		require.Contains(t, string(body), "response.create")
		require.Equal(t, profile.CipherSuites, <-hellos)
		_ = conn.Close()
	}
	require.Len(t, dialer.proxyClients, 2)
	for _, entry := range dialer.proxyClients {
		entry.client.CloseIdleConnections()
	}
}
