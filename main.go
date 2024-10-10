package main

import (
	"context"
	"flag"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
	slokMiddleware "github.com/slok/go-http-metrics/middleware"
	"github.com/slok/go-http-metrics/middleware/std"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var target = flag.String("target", "https://example.com", "the target to proxy to")

func MakeProxy(target *url.URL) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(target)
	// The above proxy will send the request without modifying the Host header.
	// Since in 99% of cases, the Host reader is required for proper routing,
	// we'll get 404. Therefore, we need to fix that.
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		// It is sufficient to set just req.Host. The 'Host' header will be set automatically.
		req.Host = target.Host
	}
	return proxy
}

func main() {
	flag.Parse()

	r := chi.NewRouter()

	// If we run behind a load balancer, we want to know the real user IP address,
	// not the load balancer address.
	// TODO: figure how it works with IPv6 clients right now, verify GKE/AWS
	r.Use(middleware.RealIP)
	// RequestID and Recoverer middlewares are implicitly added by the logger.
	r.Use(httplog.RequestLogger(log.Logger))

	metricsMiddleware := slokMiddleware.New(slokMiddleware.Config{
		Recorder: metrics.NewRecorder(metrics.Config{}),
	})
	r.Use(std.HandlerProvider("", metricsMiddleware))

	url, _ := url.Parse(*target)
	proxy := MakeProxy(url)
	r.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("X-Remote-User", "vladimir")
		proxy.ServeHTTP(w, r)
	}))

	control := chi.NewRouter()
	control.Get("/ping", func(writer http.ResponseWriter, request *http.Request) {
		// GKE requires status 200, so we can't use .WriteHeader(http.StatusNoContent)
		// AWS is fine with either 200, or 204.
		writer.Write([]byte("{}"))
	})
	control.Get("/metrics", promhttp.Handler().ServeHTTP)
	srvProxy := &http.Server{Addr: ":7070", Handler: r}
	go func() {
		if err := srvProxy.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("failed to serve proxy")
		}
	}()

	srvControl := &http.Server{Addr: ":9090", Handler: control}
	go func() {
		if err := srvControl.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("failed to serve control")
		}
	}()

	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGTERM, syscall.SIGINT)
	<-sigC
	log.Info().Msg("app shutdown requested")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { srvProxy.Shutdown(ctx); wg.Done() }()
	go func() { srvControl.Shutdown(ctx); wg.Done() }()
	wg.Wait()

	log.Info().Msg("app shutdown complete")
}
