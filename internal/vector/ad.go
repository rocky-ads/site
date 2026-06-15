package vector

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/rocky-ads/site/internal/ad"
	"github.com/rocky-ads/site/internal/config"
	"github.com/rocky-ads/site/internal/currency"
	"github.com/rocky-ads/site/internal/logger"
)

var (
	ErrNoUserActivity = errors.New("no user activity to aggregate")
	adQueue           = make(chan int, config.VectorProcessingQueueSize)
)

func BuildAdEmbedding(input ad.EmbeddingInput) error {
	return BuildAdEmbeddings([]ad.EmbeddingInput{input})
}

func BuildAdEmbeddings(inputs []ad.EmbeddingInput) error {
	if len(inputs) == 0 {
		return nil
	}
	prompts := make([]string, len(inputs))
	for i, in := range inputs {
		prompts[i] = buildAdEmbeddingPrompt(in)
	}
	embeddings, err := embedTexts(prompts)
	if err != nil {
		return err
	}
	adIDs := make([]int, len(inputs))
	metas := make([]map[string]any, len(inputs))
	for i, in := range inputs {
		adIDs[i] = in.ID
		metas[i] = buildAdEmbeddingMetadata(in)
	}
	return UpsertAdEmbeddings(adIDs, embeddings, metas)
}

func buildAdEmbeddingPrompt(in ad.EmbeddingInput) string {
	var parts []string
	parts = append(parts, `Encode the following ad for semantic search.
Focus on title, description, tags, location, price, egg disputes, and field values.

Title: `+in.Title)
	parts = append(parts, "Description: "+in.Description)
	if len(in.TagLabels) > 0 {
		parts = append(parts, "Tags: "+strings.Join(in.TagLabels, ", "))
	}
	for _, d := range in.FacetDefs {
		if v, ok := in.Facets[d.Key]; ok && v.Present() {
			if s := d.EmbeddingSnippet(v); s != "" {
				parts = append(parts, s)
			}
		}
	}
	if in.City != "" || in.AdminArea != "" {
		parts = append(parts, fmt.Sprintf(
			"Location: %s, %s, %s",
			in.City, in.AdminArea, in.Country,
		))
	}
	if amount, code, ok := in.PriceValue(); ok {
		parts = append(parts, fmt.Sprintf(
			"Price: %s", currency.Format(amount, code),
		))
	}
	parts = append(parts, eggContext(in.EggCount))
	return strings.Join(parts, "\n")
}

func eggContext(count int) string {
	switch count {
	case 0:
		return "This ad has no reported disputes (0 eggs thrown)."
	case 1:
		return "This ad has 1 reported dispute (1 egg thrown)."
	default:
		return fmt.Sprintf(
			"This ad has %d reported disputes (%d eggs thrown).",
			count, count,
		)
	}
}

func buildAdEmbeddingMetadata(in ad.EmbeddingInput) map[string]any {
	meta := map[string]any{
		"category_id": in.CategoryID,
		"egg_count":   in.EggCount,
	}
	if amount, _, ok := in.PriceValue(); ok {
		meta["price"] = amount
	}
	if in.HasLocation {
		meta["location"] = map[string]any{
			"lat": in.Latitude,
			"lon": in.Longitude,
		}
	}
	for _, d := range in.FacetDefs {
		v, ok := in.Facets[d.Key]
		if !ok || !v.Present() {
			continue
		}
		if val, ok := d.VectorMetadataValue(v); ok {
			meta[d.Key] = val
		}
	}
	return meta
}

func StartBackgroundProcessor() {
	go func() {
		const chunkSize = 50
		for {
			adID := <-adQueue
			adIDs := []int{adID}
			for i := 1; i < chunkSize; i++ {
				select {
				case id := <-adQueue:
					adIDs = append(adIDs, id)
				default:
					goto processChunk
				}
			}
		processChunk:
			var inputs []ad.EmbeddingInput
			for _, id := range adIDs {
				in, err := ad.GetForEmbedding(id)
				if err != nil {
					logger.Warn("embedding load ad", "adID", id, "error", err)
					continue
				}
				inputs = append(inputs, in)
			}
			if len(inputs) == 0 {
				continue
			}
			if err := BuildAdEmbeddings(inputs); err != nil {
				logger.Error("embedding batch failed", "error", err)
			}
			time.Sleep(config.VectorProcessingSleepInterval)
		}
	}()
}

func QueueAd(adID int) {
	adQueue <- adID
}

func ProcessAdsWithoutVectors() {
	go func() {
		if err := BackfillAllAdsSync(); err != nil {
			logger.Error("backfill ad embeddings", "error", err)
		}
	}()
}

func BackfillAllAdsSync() error {
	ids, err := ad.GetAdsWithoutVectors()
	if err != nil {
		return err
	}
	const chunkSize = 50
	for i := 0; i < len(ids); i += chunkSize {
		end := i + chunkSize
		if end > len(ids) {
			end = len(ids)
		}
		var inputs []ad.EmbeddingInput
		for _, id := range ids[i:end] {
			in, err := ad.GetForEmbedding(id)
			if err != nil {
				return err
			}
			inputs = append(inputs, in)
		}
		if err := BuildAdEmbeddings(inputs); err != nil {
			return err
		}
	}
	return nil
}
