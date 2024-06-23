package middleware

import (
	"net"
	"net/http"
	"time"

	m "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
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
			logger := log.Info().
				Str("method", r.Method).
				Str("uri", r.RequestURI).
				Int("status_code", ww.Status()).
				Int("bytes", ww.BytesWritten()).
				Int64("duration", int64(time.Since(t1))).
				Str("duration_display", time.Since(t1).String()).
				Str("remote_ip", remoteIP).
				Str("proto", r.Proto)

			if len(reqID) > 0 {
				logger = logger.Str("request_id", reqID)
			}
			logger.Msg("")
		}()

		next.ServeHTTP(ww, r)
	}

	return http.HandlerFunc(fn)
}
