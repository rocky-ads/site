package imagestore

// NewDefault returns a MinIO-backed image store. MINIO_* env vars must be set.
func NewDefault() (Store, error) {
	return NewMinio()
}
