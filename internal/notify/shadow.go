package notify

import (
	"context"
	"log"
	"os"
	"sync"

	"driftcheck/internal/drift"
)

// ShadowNotifier forwards every call to the primary notifier and, in the
// background, also sends to a secondary (shadow) notifier. Errors from the
// shadow are logged but never returned to the caller. This is useful for
// testing a new notifier in production without affecting reliability.
type ShadowNotifier struct {
	primary  drift.Notifier
	shadow   drift.Notifier
	log      *log.Logger
	wg       sync.WaitGroup
}

// NewShadowNotifier creates a ShadowNotifier. primary and shadow must not be
// nil. If logger is nil, output goes to stderr.
func NewShadowNotifier(primary, shadow drift.Notifier, logger *log.Logger) *ShadowNotifier {
	if primary == nil {
		panic("shadow: primary notifier must not be nil")
	}
	if shadow == nil {
		panic("shadow: shadow notifier must not be nil")
	}
	if logger == nil {
		logger = log.New(os.Stderr, "[shadow] ", log.LstdFlags)
	}
	return &ShadowNotifier{primary: primary, shadow: shadow, log: logger}
}

// Notify sends to the primary notifier synchronously and fires the shadow
// notifier in a goroutine. The primary result is returned to the caller.
func (s *ShadowNotifier) Notify(ctx context.Context, result drift.Result) error {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := s.shadow.Notify(ctx, result); err != nil {
			s.log.Printf("shadow notifier error for service %q: %v", result.ServiceName, err)
		}
	}()
	return s.primary.Notify(ctx, result)
}

// Wait blocks until all in-flight shadow goroutines have finished. Useful in
// tests or graceful-shutdown paths.
func (s *ShadowNotifier) Wait() {
	s.wg.Wait()
}
