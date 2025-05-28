package moqt

import (
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt/moqtrace"
)

type Config struct {
	ClientSetupExtensions func() *Parameters

	ServerSetupExtensions func(req *Parameters) (rsp *Parameters, err error)

	// Configurations
	// MaxSubscribeID SubscribeID // TODO:

	// NewSessionURI string // TODO:

	Tracer func() moqtrace.SessionTracer

	// CheckRoot func(r SetupRequest) bool // TODO:

	Timeout time.Duration
}

func (c *Config) Clone() *Config {
	return &Config{
		// MaxSubscribeID: c.MaxSubscribeID,
		// NewSessionURI:  c.NewSessionURI,
		// CheckRoot:      c.CheckRoot,
		Tracer:  c.Tracer,
		Timeout: c.Timeout,
	}
}
