package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"golang.org/x/sync/errgroup"

	"github.com/kochemajaka/gofocas"
)

type snapshot struct {
	System   focas.System    `json:"system"`
	Status   focas.Status    `json:"status"`
	Axes     []focas.Axis    `json:"axes"`
	Spindles []focas.Spindle `json:"spindles"`
	Feed     focas.Feed      `json:"feed"`
	Program  focas.Program   `json:"program"`
	Alarms   []focas.Alarm   `json:"alarms"`
}

func main() {
	addr := os.Getenv("FOCAS_ADDR")
	if addr == "" {
		addr = "192.168.1.1"
	}

	ctx := context.Background()
	client, err := focas.Dial(ctx, addr)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	var s snapshot
	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error { var e error; s.System, e = client.System(gCtx); return e })
	g.Go(func() error { var e error; s.Status, e = client.Status(gCtx); return e })
	g.Go(func() error { var e error; s.Axes, e = client.Axes(gCtx); return e })
	g.Go(func() error { var e error; s.Spindles, e = client.Spindles(gCtx); return e })
	g.Go(func() error { var e error; s.Feed, e = client.Feed(gCtx); return e })
	g.Go(func() error { var e error; s.Program, e = client.ExecutingProgram(gCtx); return e })
	g.Go(func() error { var e error; s.Alarms, e = client.Alarms(gCtx); return e })

	if err := g.Wait(); err != nil {
		log.Fatalf("read: %v", err)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(s); err != nil {
		log.Fatal(err)
	}
}
