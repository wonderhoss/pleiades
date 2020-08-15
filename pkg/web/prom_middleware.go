package web

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "pleiades_web_http_duration_seconds",
		Help: "Duration of HTTP requests",
	}, []string{"path"})

	httpStatus = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "pleiades_web_http_response_total",
		Help: "Total number of HTTP responses sent",
	}, []string{"path", "status"})
)

type recordingResponseWriter struct {
	writer http.ResponseWriter
	path   string
}

func (r *recordingResponseWriter) WriteHeader(code int) {
	httpStatus.WithLabelValues(r.path, strconv.Itoa(code)).Inc()
	r.writer.WriteHeader(code)
}

func (r *recordingResponseWriter) Header() http.Header {
	return r.writer.Header()
}

func (r *recordingResponseWriter) Write(b []byte) (int, error) {
	return r.writer.Write(b)
}

func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := mux.CurrentRoute(r)
		p, _ := route.GetPathTemplate()
		timer := prometheus.NewTimer(httpDuration.WithLabelValues(p))
		rw := &recordingResponseWriter{writer: w, path: p}
		next.ServeHTTP(rw, r)
		timer.ObserveDuration()
	})
}
