package focas

import (
	"log/slog"
	"time"

	"github.com/kochemajaka/gofocas/series"
)

// Logger is the structured logging interface accepted by the client.
// It mirrors slog's method signatures so slog.Logger satisfies it via SlogLogger.
type Logger interface {
	Debug(msg string, attrs ...any)
	Info(msg string, attrs ...any)
	Warn(msg string, attrs ...any)
	Error(msg string, attrs ...any)
}

// SlogLogger wraps a *slog.Logger to satisfy Logger.
func SlogLogger(l *slog.Logger) Logger { return slogAdapter{l} }

type slogAdapter struct{ l *slog.Logger }

func (a slogAdapter) Debug(msg string, attrs ...any) { a.l.Debug(msg, attrs...) }
func (a slogAdapter) Info(msg string, attrs ...any)  { a.l.Info(msg, attrs...) }
func (a slogAdapter) Warn(msg string, attrs ...any)  { a.l.Warn(msg, attrs...) }
func (a slogAdapter) Error(msg string, attrs ...any) { a.l.Error(msg, attrs...) }

// nopLogger discards all output.
type nopLogger struct{}

func (nopLogger) Debug(_ string, _ ...any) {}
func (nopLogger) Info(_ string, _ ...any)  {}
func (nopLogger) Warn(_ string, _ ...any)  {}
func (nopLogger) Error(_ string, _ ...any) {}

// ReconnectPolicy controls how the client reacts to transient FOCAS errors.
type ReconnectPolicy struct {
	Enabled     bool
	MaxAttempts int           // 0 = unlimited
	InitialWait time.Duration // default 200ms
	MaxWait     time.Duration // default 5s
	Multiplier  float64       // default 2.0
}

// DefaultReconnectPolicy returns a sensible reconnect configuration.
func DefaultReconnectPolicy() ReconnectPolicy {
	return ReconnectPolicy{
		Enabled:     true,
		MaxAttempts: 5,
		InitialWait: 200 * time.Millisecond,
		MaxWait:     5 * time.Second,
		Multiplier:  2.0,
	}
}

type clientConfig struct {
	port        uint16
	series      series.Series
	dialTimeout time.Duration
	callTimeout time.Duration
	reconnect   ReconnectPolicy
	logger      Logger
	maxAxes     int
	maxSpindles int
	logPath     string
}

func defaultConfig() clientConfig {
	return clientConfig{
		port:        8193,
		series:      series.Unknown,
		dialTimeout: 2 * time.Second,
		callTimeout: 5 * time.Second,
		reconnect:   ReconnectPolicy{},
		logger:      nopLogger{},
		maxAxes:     0, // auto from sysinfo
		maxSpindles: 4,
		logPath:     "fanuc.log",
	}
}

// Option configures a Client at dial time.
type Option func(*clientConfig)

// WithPort overrides the FOCAS TCP port (default 8193).
func WithPort(p uint16) Option { return func(c *clientConfig) { c.port = p } }

// WithSeries forces a specific series instead of auto-detection via cnc_sysinfo.
func WithSeries(s series.Series) Option { return func(c *clientConfig) { c.series = s } }

// WithDialTimeout sets the connection timeout (default 2s).
func WithDialTimeout(d time.Duration) Option { return func(c *clientConfig) { c.dialTimeout = d } }

// WithCallTimeout sets the per-call FOCAS timeout (default 5s).
func WithCallTimeout(d time.Duration) Option { return func(c *clientConfig) { c.callTimeout = d } }

// WithReconnect enables automatic reconnection on transient errors.
func WithReconnect(p ReconnectPolicy) Option { return func(c *clientConfig) { c.reconnect = p } }

// WithLogger sets a structured logger. If not set, the client is silent.
func WithLogger(l Logger) Option { return func(c *clientConfig) { c.logger = l } }

// WithMaxAxes overrides the axis count instead of reading it from cnc_sysinfo.
func WithMaxAxes(n int) Option { return func(c *clientConfig) { c.maxAxes = n } }

// WithMaxSpindles sets the spindle count to query (default 4).
func WithMaxSpindles(n int) Option { return func(c *clientConfig) { c.maxSpindles = n } }

// WithLogPath sets the FOCAS internal log file path (default "./fanuc.log").
func WithLogPath(path string) Option { return func(c *clientConfig) { c.logPath = path } }
