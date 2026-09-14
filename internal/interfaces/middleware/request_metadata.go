package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"net"
	"net/http"

	chiMw "github.com/go-chi/chi/v5/middleware"

	"github.com/victorotene80/medilog-api/internal/shared/requestmeta"
)

func RequestMetadata(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua := r.UserAgent()

		deviceID := r.Header.Get("X-Device-ID")
		deviceName := r.Header.Get("X-Device-Name")
		acceptLang := r.Header.Get("Accept-Language")

		if deviceName == "" {
			deviceName = ua
		}

		ip := clientIP(r)

		fingerprint := r.Header.Get("X-Device-Fingerprint")
		if fingerprint == "" {
			fpRaw := ua + "|" + deviceID + "|" + acceptLang
			fpHash := sha256.Sum256([]byte(fpRaw))
			fingerprint = hex.EncodeToString(fpHash[:])
		}

		reqID := chiMw.GetReqID(r.Context())
		if reqID == "" {
			reqID = r.Header.Get("X-Request-ID")
		}

		meta := requestmeta.Meta{
			IPAddress:         ip,
			UserAgent:         ua,
			DeviceID:          deviceID,
			DeviceFingerprint: fingerprint,
			DeviceName:        deviceName,
			RequestID:         reqID,
		}

		ctx := requestmeta.WithMeta(r.Context(), meta)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return host
}
