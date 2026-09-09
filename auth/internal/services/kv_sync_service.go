package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"auth/internal/repositories"
)

type KVSyncService interface {
	SyncImpactStatsAsync()
}

type kvSyncService struct {
	userRepo repositories.UserRepository
}

func NewKVSyncService(userRepo repositories.UserRepository) KVSyncService {
	return &kvSyncService{userRepo: userRepo}
}

type ImpactUpdatePayload struct {
	ActiveParticipants int64 `json:"activeParticipants"`
	Volunteers         int64 `json:"volunteers"`
}

func (s *kvSyncService) SyncImpactStatsAsync() {
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

	payload, err := s.fetchImpactMetrics(ctx)
	if err != nil {
		log.Printf("[KVSyncService] %v", err)
		return
	}

	if err := s.sendWorkerUpdate(ctx, workerURL, syncToken, payload); err != nil {
		log.Printf("[KVSyncService] %v", err)
		return
	}

	log.Printf("[KVSyncService] Successfully updated Cloudflare Worker KV stats (Participants: %d, Volunteers: %d)", payload.ActiveParticipants, payload.Volunteers)
}

func (s *kvSyncService) fetchImpactMetrics(ctx context.Context) (*ImpactUpdatePayload, error) {
	activeCount, err := s.userRepo.CountActiveParticipants(ctx)
	if err != nil {
		return nil, fmt.Errorf("Error counting active participants: %w", err)
	}

	volunteerCount, err := s.userRepo.CountVolunteers(ctx)
	if err != nil {
		return nil, fmt.Errorf("Error counting volunteers: %w", err)
	}

	return &ImpactUpdatePayload{
		ActiveParticipants: activeCount,
		Volunteers:         volunteerCount,
	}, nil
}

func (s *kvSyncService) sendWorkerUpdate(ctx context.Context, workerURL, syncToken string, payload *ImpactUpdatePayload) error {
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("Error marshaling JSON payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", workerURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("Error creating HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if syncToken != "" {
		req.Header.Set("Authorization", "Bearer "+syncToken)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Error posting to Cloudflare Worker KV: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Worker returned status code %d when updating KV stats", resp.StatusCode)
	}

	return nil
}
