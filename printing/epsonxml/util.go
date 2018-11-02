package epsonxml

import (
	"encoding/base64"
	"strconv"

	"pos-proxy/printing"
)

// ImgToXML takes image bytes with width and height
// return Image struct base64 encoding, color `color_1` and mode `mono`
func ImgToXML(data []byte, width int, height int) *printing.Image {
	return &printing.Image{
		Image:  base64.StdEncoding.EncodeToString(data),
		Width:  strconv.Itoa(width),
		Height: strconv.Itoa(height),
		Color:  "color_1",
		Mode:   "mono",
	}

}
