package alert

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

// snsPublisher abstracts the SNS client for testing.
type snsPublisher interface {
	Publish(ctx context.Context, params *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error)
}

// SNSNotifier sends alerts to an AWS SNS topic.
type SNSNotifier struct {
	client   snsPublisher
	topicARN string
	region   string
}

// NewSNSNotifier creates an SNSNotifier targeting the given topic ARN.
// The AWS region is loaded from the environment or shared config.
func NewSNSNotifier(topicARN, region string) (*SNSNotifier, error) {
	if topicARN == "" {
		return nil, fmt.Errorf("sns: topic ARN must not be empty")
	}
	if region == "" {
		return nil, fmt.Errorf("sns: region must not be empty")
	}

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("sns: failed to load AWS config: %w", err)
	}

	return &SNSNotifier{
		client:   sns.NewFromConfig(cfg),
		topicARN: topicARN,
		region:   region,
	}, nil
}

// Send publishes the alert to the configured SNS topic.
func (n *SNSNotifier) Send(a Alert) error {
	subject := fmt.Sprintf("[VaultWatch] %s secret expiry alert", a.Level)
	body := fmt.Sprintf("%s\n\nSecret: %s\nExpires in: %s",
		a.Message, a.SecretPath, a.TimeLeft)

	_, err := n.client.Publish(context.Background(), &sns.PublishInput{
		TopicArn: aws.String(n.topicARN),
		Subject:  aws.String(subject),
		Message:  aws.String(body),
	})
	if err != nil {
		return fmt.Errorf("sns: failed to publish message: %w", err)
	}
	return nil
}
