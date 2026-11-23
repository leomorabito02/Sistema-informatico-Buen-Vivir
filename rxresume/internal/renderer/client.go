package renderer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"buenvivir/rxresume/internal/models"
	"google.golang.org/api/idtoken"
)

type Client interface {
	RenderPDF(ctx context.Context, resume *models.Resume) ([]byte, error)
}

type client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) Client {
	httpClient := &http.Client{
		Timeout: 60 * time.Second,
	}

	if strings.HasPrefix(baseURL, "https://") {
		if idClient, err := idtoken.NewClient(context.Background(), baseURL); err == nil {
			idClient.Timeout = 60 * time.Second
			httpClient = idClient
		} else {
			slog.Warn("renderer: failed to create idtoken client, falling back to default HTTP client", "error", err)
		}
	}

	return &client{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

type renderRequest struct {
	CV models.Resume `json:"cv"`
}

func (c *client) RenderPDF(ctx context.Context, resume *models.Resume) ([]byte, error) {
	payload := renderRequest{CV: *resume}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal render request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/render", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create render request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call cv-renderer: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("cv-renderer returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return io.ReadAll(resp.Body)
}
