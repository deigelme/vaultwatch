package alert

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/pubsub/pstest"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestNewGooglePubSubNotifier_EmptyProjectID(t *testing.T) {
	_, err := NewGooglePubSubNotifier("", "my-topic")
	if err == nil {
		t.Fatal("expected error for empty projectID")
	}
}

func TestNewGooglePubSubNotifier_EmptyTopicID(t *testing.T) {
	_, err := NewGooglePubSubNotifier("my-project", "")
	if err == nil {
		t.Fatal("expected error for empty topicID")
	}
}

func TestGooglePubSubNotifier_Send_Success(t *testing.T) {
	const projectID = "test-project"
	const topicID = "test-topic"

	srv := pstest.NewServer()
	t.Cleanup(func() { srv.Close() })

	conn, err := grpc.NewClient(srv.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("grpc dial: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID, option.WithGRPCConn(conn))
	if err != nil {
		t.Fatalf("pubsub client: %v", err)
	}
	t.Cleanup(func() { client.Close() })

	if _, err := client.CreateTopic(ctx, topicID); err != nil {
		t.Fatalf("create topic: %v", err)
	}

	n := &GooglePubSubNotifier{
		projectID: projectID,
		topicID:   topicID,
		client:    client,
	}

	a := Alert{
		Level:      LevelWarning,
		SecretPath: "secret/myapp/db",
		ExpiresAt:  time.Now().Add(48 * time.Hour),
		TimeLeft:   48 * time.Hour,
	}

	if err := n.Send(a); err != nil {
		t.Fatalf("Send: %v", err)
	}

	msgs := srv.Messages()
	if len(msgs) != 1 {
		t.Fatalf("expected 1 published message, got %d", len(msgs))
	}

	var got map[string]string
	if err := json.Unmarshal(msgs[0].Data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got["secret"] != "secret/myapp/db" {
		t.Errorf("secret = %q, want %q", got["secret"], "secret/myapp/db")
	}
	if got["level"] != string(LevelWarning) {
		t.Errorf("level = %q, want %q", got["level"], LevelWarning)
	}
}
