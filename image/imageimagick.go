// +build imagick

package image

import (
	"gopkg.in/gographics/imagick.v2/imagick"
	"strconv"
	"strings"
)

// ImageCfg is the Configuration variable
var ImageCfg Image

// SetImageCfg - Set package configuration from outside package
func SetImageCfg(Cfg Image) {
	ImageCfg = Cfg
}

// GetImagePath - Get the web path for the picture based on the payment hash
func GetImagePath(rHash string) string {
	return (ImageCfg.ImageURLDir + "/" + rHash + ".jpg")
}

// GetImagePath - Get the web path for the picture based on the payment hash
func GenerateImage(rHash string, Amount int64) string {
	imagick.Initialize()
	defer imagick.Terminate()
	var imerr error
	var text string

	mw := imagick.NewMagickWand()
	dw := imagick.NewDrawingWand()
	pw := imagick.NewPixelWand()

	imerr = mw.ReadImage(ImageCfg.ImageFile)
	if imerr != nil {
		panic(imerr)
	}
	pw.SetColor(ImageCfg.ImageTextColor)
	dw.SetFillColor(pw)
	dw.SetFont(ImageCfg.ImageTextFont)
	dw.SetFontSize(ImageCfg.ImageTextSize)
	dw.SetStrokeColor(pw)
	text = strings.Replace(ImageCfg.ImageText, "{Amount}", strconv.FormatInt(Amount, 10), -1)
	dw.Annotation(ImageCfg.ImageTextX, ImageCfg.ImageTextY, text)
	mw.DrawImage(dw)
	mw.WriteImage(ImageCfg.ImageDir + "/" + rHash + ".jpg")
	newWidth := uint(250)
	newHeight := uint(mw.GetImageHeight() * newWidth / mw.GetImageWidth())
	imerr = mw.ResizeImage(newWidth, newHeight, 0, 1)
	if imerr != nil {
		panic(imerr)
	}
	mw.WriteImage("/var/www/lnd/tips/" + rHash + ".jpg_small.jpg")

	return GetImagePath(rHash)
}
