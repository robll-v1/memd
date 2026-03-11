package embed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type Embedder interface {
	Enabled() bool
	BatchEmbed(ctx context.Context, texts []string) ([][]float32, error)
}

type Config struct {
	Provider string
	BaseURL  string
	APIKey   string
	Model    string
	Dims     int
}

type noopEmbedder struct{}

func (noopEmbedder) Enabled() bool { return false }
func (noopEmbedder) BatchEmbed(_ context.Context, texts []string) ([][]float32, error) {
	out := make([][]float32, len(texts))
	return out, nil
}

type openAIEmbedder struct {
	baseURL string
	apiKey  string
	model   string
	dims    int
	client  *http.Client
}

func (e *openAIEmbedder) Enabled() bool { return true }

func (e *openAIEmbedder) BatchEmbed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}
	body := map[string]any{
		"model": e.model,
		"input": texts,
	}
	if e.dims > 0 {
		body["dimensions"] = e.dims
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(e.baseURL, "/")+"/embeddings", bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if e.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+e.apiKey)
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("embedding request failed: %s", resp.Status)
	}
	var parsed struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}
	result := make([][]float32, 0, len(parsed.Data))
	for _, item := range parsed.Data {
		result = append(result, item.Embedding)
	}
	if len(result) != len(texts) {
		return nil, fmt.Errorf("embedding response count mismatch: got %d want %d", len(result), len(texts))
	}
	return result, nil
}

func New(cfg Config) Embedder {
	provider := strings.TrimSpace(strings.ToLower(cfg.Provider))
	if provider == "" {
		provider = strings.TrimSpace(strings.ToLower(os.Getenv("MEMD_EMBED_PROVIDER")))
	}
	if provider == "" || provider == "none" {
		return noopEmbedder{}
	}
	if provider != "openai" {
		return noopEmbedder{}
	}
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = os.Getenv("MEMD_EMBED_BASE_URL")
	}
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	apiKey := cfg.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("MEMD_EMBED_API_KEY")
	}
	model := cfg.Model
	if model == "" {
		model = os.Getenv("MEMD_EMBED_MODEL")
	}
	if model == "" {
		model = "text-embedding-3-small"
	}
	return &openAIEmbedder{
		baseURL: baseURL,
		apiKey:  apiKey,
		model:   model,
		dims:    cfg.Dims,
		client:  &http.Client{Timeout: 20 * time.Second},
	}
}
