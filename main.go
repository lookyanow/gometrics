package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	version = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "version",
		Help: "Version information about this binary",
		ConstLabels: map[string]string{
			"version": "v0.1.0",
		},
	})
	httpRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Count of all HTTP requests",
	}, []string{"code", "method"})

	httpTestRequestTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_test_request_total",
		Help: "Count of test HTTP requests",
	}, []string{"code", "method"})

	testMetric = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "test_metric",
		Help: "Shows test metric gauge",
	})
)

func main() {
	bind := ""
	flagset := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flagset.StringVar(&bind, "bind", ":8080", "The socket to bind to.")
	flagset.Parse(os.Args[1:])

	r := prometheus.NewRegistry()
	r.MustRegister(httpRequestsTotal)
	r.MustRegister(httpTestRequestTotal)
	r.MustRegister(version)

	r.MustRegister(testMetric)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello from example application.\n"))
	})

	notfound := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	test := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Test message\n"))
		headerType := r.Header.Get("Type")
		if headerType == "test" {
			testMetric.Inc()
		}
	})
	http.Handle("/", promhttp.InstrumentHandlerCounter(httpRequestsTotal, handler))
	http.Handle("/err", promhttp.InstrumentHandlerCounter(httpRequestsTotal, notfound))
	http.Handle("/test", promhttp.InstrumentHandlerCounter(httpTestRequestTotal, test))

	http.Handle("/metrics", promhttp.HandlerFor(r, promhttp.HandlerOpts{}))
	log.Fatal(http.ListenAndServe(bind, nil))
}
