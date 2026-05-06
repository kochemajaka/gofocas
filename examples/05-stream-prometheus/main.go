package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kochemajaka/gofocas"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	feedActual = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cnc_feed_actual_mm_per_min",
		Help: "Actual feed rate in mm/min",
	})
	feedOverride = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cnc_feed_override_percent",
		Help: "Feed override percentage",
	})
	spindleLoad = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cnc_spindle_load_percent",
		Help: "Spindle load percentage",
	}, []string{"index"})
	spindleRPM = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cnc_spindle_rpm",
		Help: "Spindle speed in RPM",
	}, []string{"index"})
)

func init() {
	prometheus.MustRegister(feedActual, feedOverride, spindleLoad, spindleRPM)
}

func collect(client *focas.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if f, err := client.Feed(ctx); err == nil {
		feedActual.Set(float64(f.ActualMMPerMin))
		feedOverride.Set(float64(f.OverridePercent))
	}
	if spindles, err := client.Spindles(ctx); err == nil {
		for _, s := range spindles {
			label := fmt.Sprintf("%d", s.Index)
			spindleLoad.WithLabelValues(label).Set(s.Load)
			spindleRPM.WithLabelValues(label).Set(float64(s.SpeedRPM))
		}
	}
}

func main() {
	addr := os.Getenv("FOCAS_ADDR")
	if addr == "" {
		addr = "192.168.1.1"
	}
	listenAddr := os.Getenv("LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = ":2112"
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	client, err := focas.Dial(ctx, addr,
		focas.WithReconnect(focas.DefaultReconnectPolicy()),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				collect(client)
			}
		}
	}()

	srv := &http.Server{Addr: listenAddr}
	http.Handle("/metrics", promhttp.Handler())
	log.Printf("Serving metrics on %s", listenAddr)

	go func() {
		<-ctx.Done()
		srv.Shutdown(context.Background())
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
