package bootstrap

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	svchttp "github.com/victorotene80/medilog-api/internal/infrastructure/services/http"
	"github.com/victorotene80/medilog-api/internal/shared/config"
)

func initializeHTTPClient(cfg config.HTTPConfig) svchttp.HTTPService {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		// Disable HTTP/2 to prevent 'bad record MAC' connection corruption with Cloudflare
		ForceAttemptHTTP2:     false,
		TLSNextProto:          make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig: &tls.Config{
			// Try to avoid TLS fingerprinting drops
			MinVersion: tls.VersionTLS12,
		},
	}

	return svchttp.NewDefaultHTTPService(&http.Client{
		Timeout:   cfg.Timeout,
		Transport: transport,
	})
}
