package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"buenvivir/rxresume/internal/client"
	"buenvivir/rxresume/internal/repositories"
	"buenvivir/rxresume/internal/storage"
	"cloud.google.com/go/pubsub"
)

type pubSubSyncQueue struct {
	client       *pubsub.Client
	topic        *pubsub.Topic
	subscription *pubsub.Subscription
	repo         repositories.ResumeRepository
	rxClient     client.ReactiveResumeClient
	storage      storage.ObjectStorageService
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
	storage storage.ObjectStorageService,
) (SyncQueue, error) {
	psClient, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}

	topic := psClient.Topic(topicID)
	exists, err := topic.Exists(ctx)
	if err == nil && !exists {
		topic, _ = psClient.CreateTopic(ctx, topicID)
	}

	subscription := psClient.Subscription(subID)
	exists, err = subscription.Exists(ctx)
	if err == nil && !exists {
		subscription, _ = psClient.CreateSubscription(ctx, subID, pubsub.SubscriptionConfig{
			Topic:       topic,
			AckDeadline: 10 * time.Second,
		})
	}

	return &pubSubSyncQueue{
		client:       psClient,
		topic:        topic,
		subscription: subscription,
		repo:         repo,
		rxClient:     rxClient,
		storage:      storage,
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
		backoff := 5 * time.Second
		for {
			if ctx.Err() != nil {
				return
			}
			
			// Configurar soporte para alta concurrencia: limitamos los mensajes simultáneos
			// para no saturar la memoria (PDF generation consume recursos considerables).
			q.subscription.ReceiveSettings.MaxOutstandingMessages = 5
			
			log.Printf("PubSub worker: listening to subscription %s", q.subscription.ID())
			err := q.subscription.Receive(ctx, q.handleMessage)
			if ctx.Err() != nil {
				log.Println("PubSub worker: context cancelled, shutting down")
				return
			}
			if err != nil {
				log.Printf("PubSub worker: subscriber stopped with error: %v. Retrying in %s", err, backoff)
				time.Sleep(backoff)
				if backoff < 60*time.Second {
					backoff *= 2
				}
			}
		}
	}()
}

func (q *pubSubSyncQueue) handleMessage(ctx context.Context, msg *pubsub.Message) {
	var syncMsg SyncMessage
	if err := json.Unmarshal(msg.Data, &syncMsg); err != nil {
		log.Printf("PubSub worker: invalid message format, dropping. %v", err)
		msg.Ack()
		return
	}

	// Validación estricta
	if syncMsg.UserID == "" {
		log.Println("PubSub worker: invalid message, UserID is empty. Dropping.")
		msg.Ack()
		return
	}

	if err := q.process(ctx, syncMsg.UserID); err != nil {
		log.Printf("PubSub worker: processing failed for user %s: %v. Nacking message for retry.", syncMsg.UserID, err)
		msg.Nack()
	} else {
		log.Printf("PubSub worker: sync successful for user %s.", syncMsg.UserID)
		msg.Ack()
	}
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

	err = q.rxClient.SyncResume(ctx, resume)
	if err != nil {
		return err
	}
	_ = q.repo.Save(ctx, resume)

	// 3. Generate PDF and upload to GCS
	pdfCtx, pdfCancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer pdfCancel()

	pdfBytes, err := q.rxClient.ExportResumePDF(pdfCtx, userID)
	if err != nil {
		log.Printf("PubSub worker: Error generating PDF for user %s: %v", userID, err)
		return nil // Return nil so it acks. PDF can be retried on next save.
	}

	if len(pdfBytes) == 0 {
		log.Printf("PubSub worker: Generated PDF for user %s is empty (0 bytes). Skipping upload.", userID)
		return nil 
	}

	objectName := fmt.Sprintf("pdfs/%s.pdf", userID)
	_, err = q.storage.UploadFile(ctx, objectName, bytes.NewReader(pdfBytes), "application/pdf")
	if err != nil {
		log.Printf("PubSub worker: Error uploading PDF to GCS for user %s: %v", userID, err)
	} else {
		log.Printf("PubSub worker: PDF generated and cached in GCS for user %s", userID)
	}

	return nil
}
