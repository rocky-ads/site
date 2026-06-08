package imagestore

// Store persists ad image files. Local disk is the default; MinIO can
// implement this interface when object storage replaces local files.
type Store interface {
	Put(adID, index int, suffix string, data []byte) error
}
