package vector

import (
	"context"
	"fmt"
	"strings"

	"github.com/rocky-ads/site/internal/config"
	genai "google.golang.org/genai"
)

type geminiEmbedder struct {
	client *genai.Client
}

func InitGeminiClient() error {
	if config.GeminiAPIKey == "" {
		return fmt.Errorf("missing GEMINI_API_KEY")
	}
	client, err := genai.NewClient(context.Background(), nil)
	if err != nil {
		return err
	}
	SetEmbedder(geminiEmbedder{client: client})
	return nil
}

func (g geminiEmbedder) EmbedText(text string) ([]float32, error) {
	vecs, err := g.EmbedTexts([]string{text})
	if err != nil || len(vecs) == 0 {
		return nil, err
	}
	return vecs[0], nil
}

func (g geminiEmbedder) EmbedTexts(texts []string) ([][]float32, error) {
	if g.client == nil {
		return nil, fmt.Errorf("Gemini client not initialized")
	}
	if len(texts) == 0 {
		return nil, fmt.Errorf("cannot embed empty text array")
	}
	var contents []*genai.Content
	for i, text := range texts {
		text = strings.TrimSpace(text)
		if text == "" {
			return nil, fmt.Errorf("cannot embed empty text at index %d", i)
		}
		contents = append(contents, genai.Text(text)...)
	}
	dim := int32(config.GeminiEmbeddingDimensions)
	resp, err := g.client.Models.EmbedContent(
		context.Background(),
		config.GeminiEmbeddingModel,
		contents,
		&genai.EmbedContentConfig{OutputDimensionality: &dim},
	)
	if err != nil {
		return nil, fmt.Errorf("Gemini embedding API: %w", err)
	}
	if resp == nil || len(resp.Embeddings) != len(texts) {
		return nil, fmt.Errorf("unexpected Gemini embedding response")
	}
	out := make([][]float32, len(texts))
	for i, emb := range resp.Embeddings {
		if emb == nil {
			return nil, fmt.Errorf("nil embedding at index %d", i)
		}
		out[i] = emb.Values
	}
	return out, nil
}
