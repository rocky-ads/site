package handler

import (
	"math"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/internal/ad"
	"github.com/rocky-ads/site/internal/local"
	"github.com/rocky-ads/site/internal/logger"
	"github.com/rocky-ads/site/internal/ui"
	"github.com/rocky-ads/site/internal/vector"
)

func AdminEmbeddingsHandler(c *fiber.Ctx) error {
	currentUserID := local.GetUserID(c)
	data, err := embeddingAdminData()
	if err != nil {
		logger.Error("Failed to load embedding admin data", "error", err)
		return showError(c, "Failed to load embedding data")
	}
	return render(c, ui.AdminDashboardContainerWithEmbeddings(
		"embeddings", nil, "", "", currentUserID, data,
	))
}

func AdminEmbeddingsBackfillHandler(c *fiber.Ctx) error {
	vector.TriggerBackfill()
	return AdminEmbeddingsHandler(c)
}

func AdminEmbeddingsClearQueryCacheHandler(c *fiber.Ctx) error {
	vector.ClearQueryEmbeddingCache()
	return AdminEmbeddingsHandler(c)
}

func AdminEmbeddingsClearUserCacheHandler(c *fiber.Ctx) error {
	vector.ClearUserEmbeddingCache()
	return AdminEmbeddingsHandler(c)
}

func AdminEmbeddingsClearSiteCacheHandler(c *fiber.Ctx) error {
	vector.ClearSiteEmbeddingCache()
	return AdminEmbeddingsHandler(c)
}

func AdminEmbeddingsClearAllCachesHandler(c *fiber.Ctx) error {
	vector.ClearAllEmbeddingCaches()
	return AdminEmbeddingsHandler(c)
}

func embeddingAdminData() (ui.EmbeddingAdminData, error) {
	stats, err := ad.GetEmbeddingStats()
	if err != nil {
		return ui.EmbeddingAdminData{}, err
	}
	missing, err := ad.ListMissingEmbeddings(25)
	if err != nil {
		return ui.EmbeddingAdminData{}, err
	}
	cacheStats := vector.GetEmbeddingCacheStats()
	return ui.EmbeddingAdminData{
		EmbeddedCount: stats.Embedded,
		MissingCount:  stats.Missing,
		QueueDepth:    vector.QueueDepth(),
		Caches: []ui.EmbeddingCachePanel{
			cachePanelFromStats("Query", cacheStats["query"]),
			cachePanelFromStats("User", cacheStats["user"]),
			cachePanelFromStats("Site", cacheStats["site"]),
		},
		MissingAds: missingEmbeddingRows(missing),
	}, nil
}

func cachePanelFromStats(
	name string, stats map[string]interface{},
) ui.EmbeddingCachePanel {
	if stats == nil {
		return ui.EmbeddingCachePanel{Name: name}
	}
	return ui.EmbeddingCachePanel{
		Name:       name,
		Hits:       statsInt64(stats["hits"]),
		Misses:     statsInt64(stats["misses"]),
		HitRatePct: statsFloat64(stats["hit_rate"]),
		ItemCount:  statsInt64(stats["current_items"]),
		MemoryKB:   statsFloat64(stats["memory_used_kb"]),
	}
}

func missingEmbeddingRows(
	rows []ad.MissingEmbedding,
) []ui.MissingEmbeddingRow {
	out := make([]ui.MissingEmbeddingRow, len(rows))
	for i, r := range rows {
		out[i] = ui.MissingEmbeddingRow{
			AdID:         r.ID,
			Title:        r.Title,
			CategoryName: r.CategoryName,
		}
	}
	return out
}

func statsInt64(v any) int64 {
	switch n := v.(type) {
	case int:
		return int64(n)
	case int64:
		return n
	case uint64:
		return int64(n)
	case float64:
		return int64(n)
	default:
		return 0
	}
}

func statsFloat64(v any) float64 {
	switch n := v.(type) {
	case float64:
		if math.IsNaN(n) {
			return 0
		}
		return n
	case float32:
		return float64(n)
	case int:
		return float64(n)
	case int64:
		return float64(n)
	case uint64:
		return float64(n)
	default:
		return 0
	}
}
