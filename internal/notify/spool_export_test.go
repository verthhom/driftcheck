package notify

// SpoolLen exposes the internal spool length for white-box testing.
func SpoolLen(s *SpoolNotifier) int { return len(s.spool) }

// SpoolInner exposes the inner notifier for white-box testing.
func SpoolInner(s *SpoolNotifier) Notifier { return s.inner }
