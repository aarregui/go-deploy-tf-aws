package middleware

import (
	"log/slog"
	"net"
	"net/http"
	"time"

	m "github.com/go-chi/chi/v5/middleware"
)

type LoggerFields struct {
	RequestID       string
	Method          string
	URI             string
	StatusCode      int
	Bytes           int
	Duration        int64
	DurationDisplay string
	RemoteIp        string
	Proto           string
}

func Logger(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		reqID := m.GetReqID(r.Context())
		ww := m.NewWrapResponseWriter(w, r.ProtoMajor)
		t1 := time.Now()
		defer func() {
			remoteIP, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				remoteIP = r.RemoteAddr
			}
			attrs := []slog.Attr{
				slog.String("method", r.Method),
				slog.String("uri", r.RequestURI),
				slog.Int("status_code", ww.Status()),
				slog.Int("bytes", ww.BytesWritten()),
				slog.Duration("duration", time.Since(t1)),
				slog.String("duration_display", time.Since(t1).String()),
				slog.String("remote_ip", remoteIP),
				slog.String("proto", r.Proto),
			}

			if reqID != "" {
				attrs = append(attrs, slog.String("request_id", reqID))
			}

			slog.LogAttrs(r.Context(), slog.LevelInfo, "", attrs...)
		}()

		next.ServeHTTP(ww, r)
	}

	return http.HandlerFunc(fn)
}
