package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kochemajaka/gofocas"
)

func main() {
	addr := os.Getenv("FOCAS_ADDR")
	if addr == "" {
		addr = "192.168.1.1"
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	client, err := focas.Dial(ctx, addr)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		axes, err := client.Axes(ctx)
		if err != nil {
			log.Printf("axes: %v", err)
			continue
		}
		spindles, err := client.Spindles(ctx)
		if err != nil {
			log.Printf("spindles: %v", err)
		}

		fmt.Printf("\n--- %s ---\n", time.Now().Format("15:04:05"))
		for _, a := range axes {
			fmt.Printf("  %s  abs=%.3f  mach=%.3f  load=%.1f%%\n",
				a.Name, a.Position.Absolute, a.Position.Machine, a.Load)
		}
		for _, s := range spindles {
			fmt.Printf("  spindle[%d]  rpm=%d  load=%.1f%%\n",
				s.Index, s.SpeedRPM, s.Load)
		}
	}
}
