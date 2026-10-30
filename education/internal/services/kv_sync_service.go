package services

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"buenvivir/education/internal/repositories"
)

type KVSyncService interface {
	SyncCourseStatsAsync()
}

type kvSyncService struct {
	courseRepo repositories.CourseRepository
}

func NewKVSyncService(courseRepo repositories.CourseRepository) KVSyncService {
	return &kvSyncService{courseRepo: courseRepo}
}

type CourseImpactUpdatePayload struct {
	Courses int `json:"courses"`
}

func (s *kvSyncService) SyncCourseStatsAsync() {
	if s == nil || s.courseRepo == nil {
		return
	}
	workerURL := os.Getenv("WORKER_IMPACT_URL")
	if workerURL == "" {
		return
	}
	syncToken := os.Getenv("WORKER_SYNC_TOKEN")

	go s.performSync(workerURL, syncToken)
}

func (s *kvSyncService) performSync(workerURL, syncToken string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	courses, err := s.courseRepo.List(ctx)
	if err != nil {
		log.Printf("[Education KVSyncService] Error listing courses for sync: %v", err)
		return
	}

	payload := CourseImpactUpdatePayload{
		Courses: len(courses),
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[Education KVSyncService] Error marshaling JSON payload: %v", err)
		return
	}

	// #nosec G704 -- workerURL is an internal endpoint from configuration
	req, err := http.NewRequestWithContext(ctx, "POST", workerURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Printf("[Education KVSyncService] Error creating HTTP request: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if syncToken != "" {
		req.Header.Set("Authorization", "Bearer "+syncToken)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	// #nosec G704 -- outbound POST to configured internal Cloudflare worker
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[Education KVSyncService] Error posting to Cloudflare Worker KV: %v", err)
		return
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("[Education KVSyncService] Worker returned status code %d when updating KV stats", resp.StatusCode)
		return
	}

	log.Printf("[Education KVSyncService] Successfully updated Cloudflare Worker KV stats (Courses: %d)", payload.Courses)
}
