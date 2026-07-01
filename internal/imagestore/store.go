package imagestore

// ImageRef identifies one stored ad image variant.
type ImageRef struct {
	Index  int
	Suffix string
}

// Store persists ad image files in MinIO. LocalStore exists for tests only.
type Store interface {
	Put(adID, index int, suffix string, data []byte) error
	Get(adID, index int, suffix string) ([]byte, error)
	ListAd(adID int) ([]ImageRef, error)
	DeleteAd(adID int) error
}
