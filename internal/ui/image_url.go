package ui

// adImageURLFunc returns a browser-loadable URL for an ad image, or "" if missing.
var adImageURLFunc func(adID, index int, size string) string

// SetAdImageURLFunc registers the resolver used by image templates.
func SetAdImageURLFunc(f func(adID, index int, size string) string) {
	adImageURLFunc = f
}

// AdImageSrc returns a reused PresignGet URL, or "" when the object is missing.
func AdImageSrc(adID, index int, size string) string {
	if adImageURLFunc == nil {
		return ""
	}
	return adImageURLFunc(adID, index, size)
}
