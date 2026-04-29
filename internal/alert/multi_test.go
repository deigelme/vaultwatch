package alert

import (
	"errors"
	"testing"
	"time"
)

type mockNotifier struct {
	called bool
	returnErr error
}

func (m *mockNotifier) Send(_ Alert) error {
	m.called = true
	return m.returnErr
}

func TestMultiNotifier_AllSucceed(t *testing.T) {
	a := &mockNotifier{}
	b := &mockNotifier{}
	multi := NewMultiNotifier(a, b)

	alert := Alert{
		Level:      LevelWarning,
		SecretPath: "secret/test",
		Expiry:     time.Now().Add(72 * time.Hour),
		TimeLeft:   72 * time.Hour,
	}

	if err := multi.Send(alert); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !a.called {
		t.Error("expected notifier A to be called")
	}
	if !b.called {
		t.Error("expected notifier B to be called")
	}
}

func TestMultiNotifier_OneFailsStillCallsOthers(t *testing.T) {
	failing := &mockNotifier{returnErr: errors.New("send failed")}
	succeeding := &mockNotifier{}
	multi := NewMultiNotifier(failing, succeeding)

	alert := Alert{
		Level:      LevelCritical,
		SecretPath: "secret/critical",
		Expiry:     time.Now().Add(1 * time.Hour),
		TimeLeft:   1 * time.Hour,
	}

	err := multi.Send(alert)
	if err == nil {
		t.Fatal("expected error when a notifier fails")
	}
	if !succeeding.called {
		t.Error("expected succeeding notifier to still be called")
	}
}

func TestMultiNotifier_Empty(t *testing.T) {
	multi := NewMultiNotifier()
	alert := Alert{
		Level:      LevelInfo,
		SecretPath: "secret/noop",
		Expiry:     time.Now().Add(240 * time.Hour),
		TimeLeft:   240 * time.Hour,
	}
	if err := multi.Send(alert); err != nil {
		t.Fatalf("expected no error for empty multi notifier, got: %v", err)
	}
}
