package zd_kafka


import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
)

var (
	_ context.Context = (*signalContext)(nil)
)

type signalContext struct {
	context.Context

	mu      sync.RWMutex
	err     error
	sigChan chan os.Signal
}

func WithSignals(ctx context.Context, sig ...os.Signal) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	sigCtx := newSignalContext(ctx)
	signal.Notify(sigCtx.sigChan, sig...)

	go func() {
		defer signal.Stop(sigCtx.sigChan)
		for {
			select {
			case <-sigCtx.Done():
				return
			case sig := <-sigCtx.sigChan:
				sigCtx.recvSignal(sig)
				cancel()
			}
		}
	}()

	return sigCtx
}

func newSignalContext(ctx context.Context) *signalContext {
	return &signalContext{
		Context: ctx,
		sigChan: make(chan os.Signal, 1),
	}
}

func (c *signalContext) Err() error {
	c.mu.RLock()
	err := c.err
	c.mu.RUnlock()

	if err == nil {
		err = c.Context.Err()
	}
	return err
}

func (c *signalContext) recvSignal(sig os.Signal) {
	c.mu.RLock()
	if c.err != nil {
		c.mu.RUnlock()
		return
	}
	c.mu.RUnlock()

	c.mu.Lock()
	// double check
	if c.err == nil {
		c.err = &sigErr{sig}
	}
	c.mu.Unlock()
}

type sigErr struct {
	os.Signal
}

func (s *sigErr) Error() string {
	return fmt.Sprintf("received signal: %s", s.String())
}

