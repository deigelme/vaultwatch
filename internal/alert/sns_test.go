package alert

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sns"
)

// mockSNSPublisher implements snsPublisher for tests.
type mockSNSPublisher struct {
	publishFn func(ctx context.Context, params *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error)
}

func (m *mockSNSPublisher) Publish(ctx context.Context, params *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error) {
	return m.publishFn(ctx, params, optFns...)
}

func TestNewSNSNotifier_EmptyTopicARN(t *testing.T) {
	_, err := NewSNSNotifier("", "us-east-1")
	if err == nil {
		t.Fatal("expected error for empty topic ARN, got nil")
	}
}

func TestNewSNSNotifier_EmptyRegion(t *testing.T) {
	_, err := NewSNSNotifier("arn:aws:sns:us-east-1:123456789012:my-topic", "")
	if err == nil {
		t.Fatal("expected error for empty region, got nil")
	}
}

func TestSNSNotifier_Send_Success(t *testing.T) {
	var capturedInput *sns.PublishInput

	notifier := &SNSNotifier{
		topicARN: "arn:aws:sns:us-east-1:123456789012:vaultwatch",
		region:   "us-east-1",
		client: &mockSNSPublisher{
			publishFn: func(_ context.Context, params *sns.PublishInput, _ ...func(*sns.Options)) (*sns.PublishOutput, error) {
				capturedInput = params
				return &sns.PublishOutput{}, nil
			},
		},
	}

	a := Alert{
		Level:      LevelWarning,
		SecretPath: "secret/db/password",
		Message:    "Secret expiring soon",
		TimeLeft:   48 * time.Hour,
	}

	if err := notifier.Send(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedInput == nil {
		t.Fatal("expected Publish to be called")
	}
	if *capturedInput.TopicArn != notifier.topicARN {
		t.Errorf("expected topic ARN %q, got %q", notifier.topicARN, *capturedInput.TopicArn)
	}
}

func TestSNSNotifier_Send_PublishError(t *testing.T) {
	notifier := &SNSNotifier{
		topicARN: "arn:aws:sns:us-east-1:123456789012:vaultwatch",
		region:   "us-east-1",
		client: &mockSNSPublisher{
			publishFn: func(_ context.Context, _ *sns.PublishInput, _ ...func(*sns.Options)) (*sns.PublishOutput, error) {
				return nil, errors.New("sns: access denied")
			},
		},
	}

	a := Alert{
		Level:      LevelCritical,
		SecretPath: "secret/api/key",
		Message:    "Secret critically close to expiry",
		TimeLeft:   2 * time.Hour,
	}

	if err := notifier.Send(a); err == nil {
		t.Fatal("expected error from failed Publish, got nil")
	}
}
