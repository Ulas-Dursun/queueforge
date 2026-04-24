package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ulasdursun/queueforge/internal/domain"
)

// URLFetchPayload is the expected input for URL_FETCH jobs.
type URLFetchPayload struct {
	URL string `json:"url"`
}

// URLFetchResult is the output stored after processing.
type URLFetchResult struct {
	URL       string `json:"url"`
	WordCount int    `json:"word_count"`
	FetchedAt string `json:"fetched_at"`
}

// Process executes the job based on its type.
// Returns result JSON string or an error.
func Process(ctx context.Context, job *domain.Job) (string, error) {
	switch job.Type {
	case "URL_FETCH":
		return processFetch(ctx, job.Payload)
	default:
		return "", domain.ErrInvalidJobType
	}
}

func processFetch(ctx context.Context, payload string) (string, error) {
	var p URLFetchPayload
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		return "", fmt.Errorf("invalid payload: %w", err)
	}

	if p.URL == "" {
		return "", domain.ErrInvalidPayload
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.URL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch url: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	wordCount := len(strings.Fields(string(body)))

	result := URLFetchResult{
		URL:       p.URL,
		WordCount: wordCount,
		FetchedAt: time.Now().UTC().Format(time.RFC3339),
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(resultJSON), nil
}
