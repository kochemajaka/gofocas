// Package focas provides a Go client for FANUC CNC controllers via the FOCAS
// Ethernet library (fwlib32). It supports reading machine status, axis positions,
// spindle data, alarms, programs, and production parameters across the 0i, 15,
// 16i, 18i, 30i, 31i, and 32i series.
//
// The library requires the proprietary fwlib32 SDK (header + shared/static
// library) supplied by FANUC. See docs/fwlib32-setup.md for installation.
// Without the SDK the package still compiles (stub mode) and all methods
// return ErrUnsupported; use build tag focas_cgo to enable the real binding.
//
// Quick start:
//
//	client, err := focas.Dial(ctx, "192.168.1.1")
//	if err != nil { log.Fatal(err) }
//	defer client.Close()
//
//	sys, err := client.System(ctx)
package focas
