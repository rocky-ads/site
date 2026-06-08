package handler

import "github.com/rocky-ads/site/internal/imagestore"

var adImageStore imagestore.Store = imagestore.NewLocal("static/images/ad")

// SetAdImageStore replaces the default local image store (for tests or MinIO).
func SetAdImageStore(store imagestore.Store) {
	adImageStore = store
}
