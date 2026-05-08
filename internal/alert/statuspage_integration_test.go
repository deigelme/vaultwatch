//go:build integration
// +build integration

package alert

import (
	"os"
	"testing"
	"time"
)

// TestStatusPageNotifier_Integration_Send exercises the notifier against the
// real Atlassian Statuspage API. Set the following environment variables before
// running:
//
//	STATUSPAGE_PAGE_ID
//	STATUSPAGE_COMPONENT_ID
//	STATUSPAGE_API_KEY
//
// Run with:
//
//	go test -tags integration ./internal/alert/ -run TestStatusPageNotifier_Integration_Send -v
func TestStatusPageNotifier_Integration_Send(t *testing.T) {
	pageID := os.Getenv("STATUSPAGE_PAGE_ID")
	componentID := os.Getenv("STATUSPAGE_COMPONENT_ID")
	apiKey := os.Getenv("STATUSPAGE_API_KEY")

	if pageID == "" || componentID == "" || apiKey == "" {
		t.Skip("STATUSPAGE_PAGE_ID, STATUSPAGE_COMPONENT_ID and STATUSPAGE_API_KEY must be set")
	}

	n, err := NewStatusPageNotifier(pageID, componentID, apiKey)
	if err != nil {
		t.Fatalf("NewStatusPageNotifier: %v", err)
	}

	a := Alert{
		SecretPath: "secret/data/test",
		Level:      LevelWarning,
		ExpiresAt:  time.Now().Add(48 * time.Hour),
		TimeLeft:   48 * time.Hour,
		Message:    "integration test: secret expiring soon",
	}

	if err := n.Send(a); err != nil {
		t.Fatalf("Send: %v", err)
	}
	t.Log("alert sent successfully")
}
