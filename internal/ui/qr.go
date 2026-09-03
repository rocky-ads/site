package ui

import (
	"encoding/base64"

	"github.com/skip2/go-qrcode"
)

func qrPNGDataURI(content string, px int) string {
	if content == "" {
		return ""
	}
	png, err := qrcode.Encode(content, qrcode.Medium, px)
	if err != nil {
		return ""
	}
	return "data:image/png;base64," +
		base64.StdEncoding.EncodeToString(png)
}
