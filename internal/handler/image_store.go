package handler

import (
	"github.com/rocky-ads/site/internal/config"
	"github.com/rocky-ads/site/internal/imagestore"
	"github.com/rocky-ads/site/internal/ui"
)

var adImageStore imagestore.Store
var adImageURLCache *imagestore.URLCache

// SetAdImageStore replaces the image store (required at startup; for tests use local store).
func SetAdImageStore(store imagestore.Store) {
	adImageStore = store
	urlCache, err := imagestore.NewURLCache(store)
	if err != nil {
		panic("image URL cache: " + err.Error())
	}
	adImageURLCache = urlCache
	ui.SetAdImageURLFunc(resolveAdImageURL)
}

func resolveAdImageURL(adID, index int, size string) string {
	if adImageStore == nil || adImageURLCache == nil {
		return ""
	}
	ok, err := adImageStore.Stat(adID, index, size)
	if err != nil || !ok {
		return ""
	}
	url, err := adImageURLCache.ReusedGetURL(adID, index, size,
		config.MinIOPresignedGetExpiry)
	if err != nil {
		return ""
	}
	return url
}
