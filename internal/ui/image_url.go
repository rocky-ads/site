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

var userAccountPictureURLFunc func(userID int) string

// SetUserAccountPictureURLFunc registers the resolver for account pictures.
func SetUserAccountPictureURLFunc(f func(userID int) string) {
	userAccountPictureURLFunc = f
}

// UserAccountPictureSrc returns a PresignGet URL, or "" when missing.
func UserAccountPictureSrc(userID int) string {
	if userAccountPictureURLFunc == nil {
		return ""
	}
	return userAccountPictureURLFunc(userID)
}
