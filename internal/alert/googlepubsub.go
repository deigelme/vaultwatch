package alert

import (
	"context"
	"encoding/json"
	"fmt"

	"cloud.google.com/go/pubsub"
)

// GooglePubSubNotifier publishes alert messages to a Google Cloud Pub/Sub topic.
type GooglePubSubNotifier struct {
	projectID string
	topicID   string
	client    *pubsub.Client
}

// NewGooglePubSubNotifier creates a new GooglePubSubNotifier.
// projectID and topicID must be non-empty.
func NewGooglePubSubNotifier(projectID, topicID string) (*GooglePubSubNotifier, error) {
	if projectID == "" {
		return nil, fmt.Errorf("googlepubsub: projectID must not be empty")
	}
	if topicID == "" {
		return nil, fmt.Errorf("googlepubsub: topicID must not be empty")
	}
	client, err := pubsub.NewClient(context.Background(), projectID)
	if err != nil {
		return nil, fmt.Errorf("googlepubsub: failed to create client: %w", err)
	}
	return &GooglePubSubNotifier{
		projectID: projectID,
		topicID:   topicID,
		client:    client,
	}, nil
}

// Send publishes the alert to the configured Pub/Sub topic.
func (n *GooglePubSubNotifier) Send(a Alert) error {
	payload := map[string]string{
		"level":      string(a.Level),
		"secret":     a.SecretPath,
		"message":    a.String(),
		"expires_at": a.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("googlepubsub: failed to marshal payload: %w", err)
	}
	topic := n.client.Topic(n.topicID)
	ctx := context.Background()
	result := topic.Publish(ctx, &pubsub.Message{Data: data})
	if _, err := result.Get(ctx); err != nil {
		return fmt.Errorf("googlepubsub: publish failed: %w", err)
	}
	return nil
}
