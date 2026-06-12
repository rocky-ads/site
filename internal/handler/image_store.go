package handler

import "github.com/rocky-ads/site/internal/imagestore"

var adImageStore imagestore.Store

// SetAdImageStore replaces the image store (required at startup; for tests use local store).
func SetAdImageStore(store imagestore.Store) {
	adImageStore = store
}
