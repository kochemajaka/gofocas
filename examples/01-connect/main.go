package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/kochemajaka/gofocas"
)

func main() {
	addr := os.Getenv("FOCAS_ADDR")
	if addr == "" {
		addr = "192.168.1.1"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := focas.Dial(ctx, addr,
		focas.WithDialTimeout(3*time.Second),
	)
	if err != nil {
		log.Fatalf("dial: %v", err)
	}
	defer client.Close()

	if err := client.Ping(ctx); err != nil {
		log.Fatalf("ping: %v", err)
	}

	sys, err := client.System(ctx)
	if err != nil {
		log.Fatalf("system: %v", err)
	}

	fmt.Printf("Connected to %s\n", client.Addr())
	fmt.Printf("Series:  %s\n", sys.Series)
	fmt.Printf("Version: %s\n", sys.Version)
	fmt.Printf("Axes:    %d / %d\n", sys.Axes, sys.MaxAxes)
}
