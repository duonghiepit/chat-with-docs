package metrics

import (
    "net/http"

    prom "github.com/prometheus/client_golang/prometheus"
    promhttp "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    RequestsTotal = prom.NewCounterVec(prom.CounterOpts{
        Name: "api_requests_total",
        Help: "Total API requests",
    }, []string{"endpoint", "status"})

    RequestLatencyMs = prom.NewHistogramVec(prom.HistogramOpts{
        Name: "api_request_latency_ms",
        Help: "Latency per endpoint in ms",
        Buckets: prom.LinearBuckets(50, 50, 20),
    }, []string{"endpoint"})
)

func init() {
    prom.MustRegister(RequestsTotal, RequestLatencyMs)
}

func Handler() http.Handler { return promhttp.Handler() }


