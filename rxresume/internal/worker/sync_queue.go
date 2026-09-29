package worker

import (
	"context"
	"log"

	"buenvivir/rxresume/internal/client"
	"buenvivir/rxresume/internal/repositories"
)

type SyncQueue interface {
	Enqueue(userID string)
	Start(ctx context.Context)
}

type syncQueue struct {
	queue    chan string
	repo     repositories.ResumeRepository
	rxClient client.ReactiveResumeClient
}

func NewSyncQueue(repo repositories.ResumeRepository, rxClient client.ReactiveResumeClient) SyncQueue {
	return &syncQueue{
		queue:    make(chan string, 1000), // Buffer for 1000 sync jobs
		repo:     repo,
		rxClient: rxClient,
	}
}

func (q *syncQueue) Enqueue(userID string) {
	select {
	case q.queue <- userID:
		// Successfully enqueued
	default:
		// Queue is full, log warning
		log.Printf("Warning: Sync queue is full, dropping sync job for user %s", userID)
	}
}

func (q *syncQueue) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("Sync queue shutting down")
				return
			case userID := <-q.queue:
				q.process(ctx, userID)
			}
		}
	}()
}

func (q *syncQueue) process(ctx context.Context, userID string) {
	// 1. Fetch latest resume from DB
	resume, err := q.repo.FindByUserID(ctx, userID)
	if err != nil {
		log.Printf("Sync worker: Error fetching resume for user %s: %v", userID, err)
		return
	}
	if resume == nil {
		return
	}

	// 2. Sync with RxResume API
	err = q.rxClient.SyncResume(ctx, resume)
	if err != nil {
		log.Printf("Sync worker: Error syncing resume for user %s to RxResume: %v", userID, err)
		// Here a DLQ (Dead Letter Queue) or retry mechanism could be implemented for production
	} else {
		log.Printf("Sync worker: Successfully synced resume for user %s", userID)
	}
}
