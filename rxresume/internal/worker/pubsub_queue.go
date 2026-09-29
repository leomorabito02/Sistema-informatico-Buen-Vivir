package worker

import (
	"context"
	"encoding/json"
	"log"

	"buenvivir/rxresume/internal/client"
	"buenvivir/rxresume/internal/repositories"
	"cloud.google.com/go/pubsub"
)

type pubSubSyncQueue struct {
	client       *pubsub.Client
	topic        *pubsub.Topic
	subscription *pubsub.Subscription
	repo         repositories.ResumeRepository
	rxClient     client.ReactiveResumeClient
}

type SyncMessage struct {
	UserID string `json:"user_id"`
}

func NewPubSubSyncQueue(
	ctx context.Context,
	projectID string,
	topicID string,
	subID string,
	repo repositories.ResumeRepository,
	rxClient client.ReactiveResumeClient,
) (SyncQueue, error) {
	psClient, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}

	topic := psClient.Topic(topicID)
	subscription := psClient.Subscription(subID)

	return &pubSubSyncQueue{
		client:       psClient,
		topic:        topic,
		subscription: subscription,
		repo:         repo,
		rxClient:     rxClient,
	}, nil
}

func (q *pubSubSyncQueue) Enqueue(userID string) {
	ctx := context.Background()
	msg := &SyncMessage{UserID: userID}
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("PubSub worker: failed to marshal message for user %s: %v", userID, err)
		return
	}

	result := q.topic.Publish(ctx, &pubsub.Message{
		Data: data,
	})

	// Fire and forget, but log errors if any occur asynchronously
	go func() {
		_, err := result.Get(ctx)
		if err != nil {
			log.Printf("PubSub worker: failed to publish sync job for user %s: %v", userID, err)
		}
	}()
}

func (q *pubSubSyncQueue) Start(ctx context.Context) {
	go func() {
		log.Printf("PubSub worker: listening to subscription %s", q.subscription.ID())
		err := q.subscription.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
			var syncMsg SyncMessage
			if err := json.Unmarshal(msg.Data, &syncMsg); err != nil {
				log.Printf("PubSub worker: invalid message format, dropping. %v", err)
				msg.Ack() // Ack bad messages so they don't block
				return
			}

			// Process
			err := q.process(ctx, syncMsg.UserID)
			if err != nil {
				log.Printf("PubSub worker: processing failed for user %s: %v. Nacking message for retry.", syncMsg.UserID, err)
				msg.Nack()
			} else {
				log.Printf("PubSub worker: sync successful for user %s.", syncMsg.UserID)
				msg.Ack()
			}
		})
		if err != nil {
			log.Printf("PubSub worker: subscriber stopped with error: %v", err)
		}
	}()
}

func (q *pubSubSyncQueue) process(ctx context.Context, userID string) error {
	// 1. Fetch latest resume from DB
	resume, err := q.repo.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if resume == nil {
		log.Printf("PubSub worker: Resume not found for user %s", userID)
		return nil // No error to retry, resume might be deleted
	}

	// 2. Sync with RxResume API
	err = q.rxClient.SyncResume(ctx, resume)
	if err != nil {
		return err
	}

	return nil
}
