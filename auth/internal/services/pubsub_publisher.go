package services

import (
	"context"
	"fmt"
	"sync"

	"cloud.google.com/go/pubsub/v2"
	"google.golang.org/api/option"
)

// PubSubPublisher implements EventPublisher using cloud.google.com/go/pubsub/v2.
type PubSubPublisher struct {
	client     *pubsub.Client
	mu         sync.Mutex
	publishers map[string]*pubsub.Publisher
}

// NewPubSubPublisher initializes a new PubSubPublisher connected to GCP Pub/Sub.
func NewPubSubPublisher(ctx context.Context, projectID string, opts ...option.ClientOption) (*PubSubPublisher, error) {
	if projectID == "" {
		return nil, fmt.Errorf("projectID es obligatorio")
	}
	client, err := pubsub.NewClient(ctx, projectID, opts...)
	if err != nil {
		return nil, fmt.Errorf("error creando cliente de pubsub: %w", err)
	}
	return &PubSubPublisher{
		client:     client,
		publishers: make(map[string]*pubsub.Publisher),
	}, nil
}

// Publish sends a message to the specified topic.
func (p *PubSubPublisher) Publish(ctx context.Context, topic string, msg *pubsub.Message) error {
	p.mu.Lock()
	pub, ok := p.publishers[topic]
	if !ok {
		pub = p.client.Publisher(topic)
		pub.EnableMessageOrdering = true
		p.publishers[topic] = pub
	}
	p.mu.Unlock()

	res := pub.Publish(ctx, msg)
	_, err := res.Get(ctx)
	return err
}

