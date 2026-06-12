package imagestore

// Store persists ad image files in MinIO. LocalStore exists for tests only.
type Store interface {
	Put(adID, index int, suffix string, data []byte) error
	Get(adID, index int, suffix string) ([]byte, error)
	DeleteAd(adID int) error
}
