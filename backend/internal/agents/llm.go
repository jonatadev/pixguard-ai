package agents

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/genai"
)

// JSONGenerator abstracts structured JSON generation for testability and Gemini SDK integration.
type JSONGenerator interface {
	GenerateJSON(ctx context.Context, model string, temperature float32, systemPrompt, userPrompt string) (string, TokenUsage, error)
}

// GeminiGenerator implements JSONGenerator using the official google.golang.org/genai SDK.
type GeminiGenerator struct {
	client *genai.Client
}

// NewGeminiGenerator constructs a Gemini API client from an API key.
func NewGeminiGenerator(ctx context.Context, apiKey string) (*GeminiGenerator, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("gemini api key is required")
	}
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("create genai client: %w", err)
	}
	return &GeminiGenerator{client: client}, nil
}

// GenerateJSON invokes Gemini Models.GenerateContent enforcing application/json output.
func (g *GeminiGenerator) GenerateJSON(
	ctx context.Context,
	model string,
	temperature float32,
	systemPrompt, userPrompt string,
) (string, TokenUsage, error) {
	temp := temperature
	cfg := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(systemPrompt, genai.RoleUser),
		Temperature:       &temp,
		ResponseMIMEType:  "application/json",
	}

	contents := []*genai.Content{
		genai.NewContentFromText(userPrompt, genai.RoleUser),
	}

	resp, err := g.client.Models.GenerateContent(ctx, model, contents, cfg)
	if err != nil {
		return "", TokenUsage{}, fmt.Errorf("gemini generate content (%s): %w", model, err)
	}

	var usage TokenUsage
	if resp.UsageMetadata != nil {
		usage = TokenUsage{
			PromptTokens:     resp.UsageMetadata.PromptTokenCount,
			CompletionTokens: resp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      resp.UsageMetadata.TotalTokenCount,
		}
	}
	return strings.TrimSpace(resp.Text()), usage, nil
}
