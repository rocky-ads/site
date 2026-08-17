package imagestore

import (
	"fmt"
	"regexp"
	"strconv"
)

const (
	ImageExt  = "jpg"
	ImageMIME = "image/jpeg"
)

var imageFilePattern = regexp.MustCompile(`^(\d+)-(\d+w)\.jpg$`)

func objectKey(adID, index int, suffix string) string {
	return fmt.Sprintf("%d/%d-%s.jpg", adID, index, suffix)
}

func userAccountObjectKey(userID int) string {
	return fmt.Sprintf("users/%d/account.jpg", userID)
}

func parseImageFileName(name string) (index int, suffix string, ok bool) {
	matches := imageFilePattern.FindStringSubmatch(name)
	if len(matches) != 3 {
		return 0, "", false
	}
	index, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, "", false
	}
	return index, matches[2], true
}
