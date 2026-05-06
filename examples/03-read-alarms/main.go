package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/kochemajaka/gofocas"
)

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

	alarms, err := client.Alarms(ctx)
	if err != nil {
		log.Fatalf("alarms: %v", err)
	}

	if len(alarms) == 0 {
		fmt.Println("No active alarms.")
		return
	}

	fmt.Printf("%d active alarm(s):\n", len(alarms))
	for _, a := range alarms {
		axis := ""
		if a.Axis > 0 {
			axis = fmt.Sprintf(" axis=%d", a.Axis)
		}
		fmt.Printf("  [%s]%s  %s\n", a.Code, axis, a.Message)
	}
}
