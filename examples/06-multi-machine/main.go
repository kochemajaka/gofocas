package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/kochemajaka/gofocas"
)

func main() {
	addrsEnv := os.Getenv("FOCAS_ADDRS")
	if addrsEnv == "" {
		addrsEnv = "192.168.1.1,192.168.1.2"
	}
	addrs := strings.Split(addrsEnv, ",")

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	clients := make([]*focas.Client, 0, len(addrs))
	for _, addr := range addrs {
		c, err := focas.Dial(ctx, strings.TrimSpace(addr),
			focas.WithReconnect(focas.DefaultReconnectPolicy()),
		)
		if err != nil {
			log.Printf("dial %s: %v", addr, err)
			continue
		}
		clients = append(clients, c)
	}
	defer func() {
		for _, c := range clients {
			c.Close()
		}
	}()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		var wg sync.WaitGroup
		for _, c := range clients {
			wg.Add(1)
			go func(cl *focas.Client) {
				defer wg.Done()
				rCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
				defer cancel()

				st, err := cl.Status(rCtx)
				if err != nil {
					log.Printf("%s status: %v", cl.Addr(), err)
					return
				}
				f, _ := cl.Feed(rCtx)
				fmt.Printf("%s  mode=%-8s  run=%-6s  feed=%d mm/min\n",
					cl.Addr(), st.Mode, st.Run, f.ActualMMPerMin)
			}(c)
		}
		wg.Wait()
	}
}
