package notify_test

import (
	"context"
	"errors"
	"testing"

	"driftcheck/internal/drift"
	"driftcheck/internal/notify"
)

func makeTagResult(service string) drift.Result {
	return drift.Result{ServiceName: service}
}

func TestTaggingNotifier_NilInnerPanics(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil inner")
		}
	}()
	notify.NewTaggingNotifier(nil, map[string]string{"env": "prod"})
}

func TestTaggingNotifier_EmptyTagsPanics(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for empty tags")
		}
	}()
	notify.NewTaggingNotifier(&captureNotifier{}, map[string]string{})
}

func TestTaggingNotifier_AnnotatesServiceName(t *testing.T) {
	t.Parallel()
	cap := &captureNotifier{}
	tn := notify.NewTaggingNotifier(cap, map[string]string{"env": "staging"})

	err := tn.Notify(context.Background(), makeTagResult("payments"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := cap.last.ServiceName
	if !containsAll(got, "payments", "env=staging") {
		t.Errorf("service name %q does not contain expected tags", got)
	}
}

func TestTaggingNotifier_DoesNotMutateOriginal(t *testing.T) {
	t.Parallel()
	cap := &captureNotifier{}
	tn := notify.NewTaggingNotifier(cap, map[string]string{"region": "us-east-1"})

	orig := makeTagResult("inventory")
	_ = tn.Notify(context.Background(), orig)

	if orig.ServiceName != "inventory" {
		t.Errorf("original result was mutated: got %q", orig.ServiceName)
	}
}

func TestTaggingNotifier_ForwardsInnerError(t *testing.T) {
	t.Parallel()
	sentinel := errors.New("inner failure")
	tn := notify.NewTaggingNotifier(&errorNotifier{err: sentinel}, map[string]string{"tier": "prod"})

	err := tn.Notify(context.Background(), makeTagResult("orders"))
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

// containsAll returns true when s contains every substring in subs.
func containsAll(s string, subs ...string) bool {
	for _, sub := range subs {
		if !contains(s, sub) {
			return false
		}
	}
	return true
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		(func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		})())
}
