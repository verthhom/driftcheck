package notify

// ShadowPrimary exposes the primary notifier for white-box testing.
func ShadowPrimary(s *ShadowNotifier) drift.Notifier { return s.primary }

// ShadowSecondary exposes the shadow notifier for white-box testing.
func ShadowSecondary(s *ShadowNotifier) drift.Notifier { return s.shadow }

// ShadowWait delegates to the internal WaitGroup so tests can drain
// background goroutines before making assertions.
func ShadowWait(s *ShadowNotifier) { s.Wait() }
