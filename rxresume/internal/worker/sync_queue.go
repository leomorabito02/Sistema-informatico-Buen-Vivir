package worker

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"time"

	"buenvivir/rxresume/internal/client"
	"buenvivir/rxresume/internal/repositories"
	"buenvivir/rxresume/internal/storage"
)

type SyncQueue interface {
	Enqueue(userID string)
	Start(ctx context.Context)
}

type syncQueue struct {
	queue    chan string
	repo     repositories.ResumeRepository
	rxClient client.ReactiveResumeClient
	storage  storage.ObjectStorageService
}

func NewSyncQueue(repo repositories.ResumeRepository, rxClient client.ReactiveResumeClient, storage storage.ObjectStorageService) SyncQueue {
	return &syncQueue{
		queue:    make(chan string, 1000), // Buffer for 1000 sync jobs
		repo:     repo,
		rxClient: rxClient,
		storage:  storage,
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
		return
	}
	
	_ = q.repo.Save(ctx, resume)
	log.Printf("Sync worker: Successfully synced resume for user %s", userID)

	// 3. Generate PDF and upload to GCS
	pdfCtx, pdfCancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer pdfCancel()

	pdfBytes, err := q.rxClient.ExportResumePDF(pdfCtx, userID)
	if err != nil {
		log.Printf("Sync worker: Error generating PDF for user %s: %v", userID, err)
		return
	}

	objectName := fmt.Sprintf("pdfs/%s.pdf", userID)
	_, err = q.storage.UploadFile(ctx, objectName, bytes.NewReader(pdfBytes), "application/pdf")
	if err != nil {
		log.Printf("Sync worker: Error uploading PDF to GCS for user %s: %v", userID, err)
	} else {
		log.Printf("Sync worker: PDF generated and cached in GCS for user %s", userID)
	}
}
