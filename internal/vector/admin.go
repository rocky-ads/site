package vector

func QueueDepth() int {
	return len(adQueue)
}

func GetEmbeddingCacheStats() map[string]map[string]interface{} {
	stats := make(map[string]map[string]interface{})
	if queryEmbeddingCache != nil {
		stats["query"] = queryEmbeddingCache.StatsCopy()
	}
	if userEmbeddingCache != nil {
		stats["user"] = userEmbeddingCache.StatsCopy()
	}
	if siteEmbeddingCache != nil {
		stats["site"] = siteEmbeddingCache.StatsCopy()
	}
	return stats
}

func ClearQueryEmbeddingCache() {
	if queryEmbeddingCache != nil {
		queryEmbeddingCache.Clear()
	}
}

func ClearUserEmbeddingCache() {
	if userEmbeddingCache != nil {
		userEmbeddingCache.Clear()
	}
}

func ClearSiteEmbeddingCache() {
	if siteEmbeddingCache != nil {
		siteEmbeddingCache.Clear()
	}
}

func ClearAllEmbeddingCaches() {
	ClearQueryEmbeddingCache()
	ClearUserEmbeddingCache()
	ClearSiteEmbeddingCache()
}

func TriggerBackfill() {
	ProcessAdsWithoutVectors()
}
