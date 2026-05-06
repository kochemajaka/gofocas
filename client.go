package focas

import (
	"context"
	"errors"
	"fmt"
	"net"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/kochemajaka/gofocas/internal/fwlib32"
	iseries "github.com/kochemajaka/gofocas/internal/series"
	"github.com/kochemajaka/gofocas/series"
)

const dispatchQueueSize = 64

type dispatchReq struct {
	fn  func(h fwlib32.Handle) error
	res chan error
}

// Client is a connection to a FANUC CNC controller.
// All methods are safe for concurrent use.
//
// All FOCAS calls are serialised through a single goroutine pinned to one OS
// thread via runtime.LockOSThread. The 31i (and similar) controllers return
// EW_HANDLE (-8) when the same handle is used from different OS threads
// concurrently, even if Go-level mutexes serialise the calls — the controller
// tracks the OS thread ID of the originating cnc_allclibhndl3 call.
type Client struct {
	mu      sync.Mutex
	handle  fwlib32.Handle
	closed  bool
	addr    string // host:port normalised
	host    string
	cfg     clientConfig
	binder  fwlib32.Binder
	strategy iseries.Strategy
	series  series.Series
	maxAxes int

	dispatch chan dispatchReq
	stopDisp chan struct{}
}

// Dial opens a FOCAS connection to addr. addr may be "host" or "host:port".
// If port is omitted, the port from WithPort (default 8193) is used.
func Dial(ctx context.Context, addr string, opts ...Option) (*Client, error) {
	cfg := defaultConfig()
	for _, o := range opts {
		o(&cfg)
	}

	host, port, err := splitAddr(addr, cfg.port)
	if err != nil {
		return nil, &Error{Op: "Dial", Msg: "invalid address: " + err.Error()}
	}

	b := fwlib32.New()
	if err := b.Startup(cfg.logPath); err != nil {
		return nil, wrapFocasErr("Dial", "startup failed", err)
	}

	c := &Client{
		addr:     fmt.Sprintf("%s:%d", host, port),
		host:     host,
		cfg:      cfg,
		binder:   b,
		dispatch: make(chan dispatchReq, dispatchQueueSize),
		stopDisp: make(chan struct{}),
	}

	// Start the OS-thread-locked dispatcher first so that Alloc and all
	// subsequent FOCAS calls share the same OS thread. cnc_allclibhndl3 binds
	// the handle to the calling OS thread; any later call from a different
	// thread returns EW_HANDLE (-8).
	go c.dispatchLoop()

	// Alloc must run on the dispatcher's OS-locked thread.
	timeoutMs := uint32(cfg.dialTimeout.Milliseconds())
	var h fwlib32.Handle
	if err := c.exec(ctx, func(_ fwlib32.Handle) error {
		var e error
		h, e = b.Alloc(host, port, timeoutMs)
		return e
	}); err != nil {
		c.stopDispatcher()
		return nil, wrapFocasErr("Dial", "connect failed", err)
	}
	c.handle = h

	// Auto-detect series unless overridden.
	if err := c.detectSeries(ctx); err != nil {
		c.stopDispatcher()
		_ = b.Free(h)
		return nil, err
	}

	return c, nil
}

// dispatchLoop runs on a single goroutine pinned to one OS thread.
// All FOCAS C-library calls must go through here.
func (c *Client) dispatchLoop() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	for {
		select {
		case <-c.stopDisp:
			return
		case req := <-c.dispatch:
			c.mu.Lock()
			h := c.handle
			c.mu.Unlock()
			req.res <- req.fn(h)
		}
	}
}

func (c *Client) stopDispatcher() {
	select {
	case c.stopDisp <- struct{}{}:
	default:
	}
}

// exec sends fn to the OS-locked dispatcher and waits for the result.
// It is the single chokepoint for every FOCAS call.
func (c *Client) exec(ctx context.Context, fn func(h fwlib32.Handle) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	res := make(chan error, 1)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case c.dispatch <- dispatchReq{fn: fn, res: res}:
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-res:
		return err
	}
}

func (c *Client) detectSeries(ctx context.Context) error {
	if c.cfg.series != series.Unknown {
		c.series = c.cfg.series
		c.strategy = iseries.For(c.series)
		return nil
	}

	var si fwlib32.SysInfo
	err := c.exec(ctx, func(h fwlib32.Handle) error {
		var e error
		si, e = c.binder.SysInfo(h)
		return e
	})
	if err != nil {
		return wrapFocasErr("Dial/cnc_sysinfo", "series detection failed", err)
	}

	raw := strings.TrimRight(string(si.Series[:]), "\x00 ")
	c.series = series.Parse(raw)
	c.strategy = iseries.For(c.series)

	if c.cfg.maxAxes == 0 {
		c.maxAxes = int(si.Axes)
		if c.maxAxes == 0 {
			c.maxAxes = int(si.MaxAxis)
		}
	} else {
		c.maxAxes = c.cfg.maxAxes
	}

	return nil
}

// redial closes the current handle and opens a new one (called by reconnect).
// Both Free and Alloc run on the dispatcher's OS-locked thread so the new
// handle is bound to the same OS thread as all subsequent FOCAS calls.
func (c *Client) redial(ctx context.Context) error {
	return c.exec(ctx, func(h fwlib32.Handle) error {
		_ = c.binder.Free(h)

		timeoutMs := uint32(c.cfg.dialTimeout.Milliseconds())
		newH, err := c.binder.Alloc(c.host, c.cfg.port, timeoutMs)
		if err != nil {
			return err
		}

		c.mu.Lock()
		c.handle = newH
		c.mu.Unlock()
		return nil
	})
}

// Close releases the FOCAS handle. Idempotent.
func (c *Client) Close() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true
	h := c.handle
	c.mu.Unlock()

	// Free must run on the same OS thread as Alloc, so route it through the
	// dispatcher before stopping it.
	ctx := context.Background()
	var freeErr error
	_ = c.exec(ctx, func(_ fwlib32.Handle) error {
		freeErr = c.binder.Free(h)
		return nil
	})
	c.stopDispatcher()
	if freeErr != nil {
		return &Error{Op: "Close", Err: freeErr}
	}
	return nil
}

// Ping performs a cheap cnc_sysinfo call to verify the connection is alive.
func (c *Client) Ping(ctx context.Context) error {
	return c.call(ctx, "Ping", func(h fwlib32.Handle) error {
		_, err := c.binder.SysInfo(h)
		return err
	})
}

// Addr returns the normalised host:port this client is connected to.
func (c *Client) Addr() string { return c.addr }

// Series returns the detected or configured series.
func (c *Client) Series() series.Series {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.series
}

// call is the public-facing chokepoint for all reader methods. It routes
// through the OS-locked dispatcher and applies the reconnect policy on error.
func (c *Client) call(ctx context.Context, op string, fn func(h fwlib32.Handle) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return ErrClosed
	}
	c.mu.Unlock()

	doFn := func() error {
		err := c.exec(ctx, fn)
		if err == nil {
			return nil
		}
		var fe *Error
		if errors.As(err, &fe) {
			return err
		}
		var fw *fwlib32.FocasError
		if errors.As(err, &fw) {
			return &Error{Op: fw.Op, Code: Code(fw.Code)}
		}
		return &Error{Op: op, Err: err}
	}

	return c.reconnectDo(ctx, doFn)
}

func splitAddr(addr string, defaultPort uint16) (host string, port uint16, err error) {
	if strings.ContainsRune(addr, ':') {
		h, p, e := net.SplitHostPort(addr)
		if e != nil {
			return "", 0, e
		}
		n, e := strconv.ParseUint(p, 10, 16)
		if e != nil {
			return "", 0, e
		}
		return h, uint16(n), nil
	}
	return addr, defaultPort, nil
}
