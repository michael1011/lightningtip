package image

import (
	"strconv"

	"gopkg.in/gographics/imagick.v2/imagick"
)

func GetImagePath(rHash string) string {
   return("/tips" + "/" + rHash + ".jpg")
}

func GenerateImage(rHash string, Amount int64) string {
  imagick.Initialize()
  defer imagick.Terminate()
  var imerr error

  mw := imagick.NewMagickWand()
  dw := imagick.NewDrawingWand()
  pw := imagick.NewPixelWand()

  imerr = mw.ReadImage("/var/www/lnd/tips/kasan.jpg")
  if imerr != nil {
      panic(imerr)
  }
  pw.SetColor("black")
  dw.SetFillColor(pw)
  dw.SetFont("Verdana-Bold-Italic")
  dw.SetFontSize(150)
  dw.SetStrokeColor(pw)
  dw.Annotation(25, 165, "I paid a random dude " + strconv.FormatInt(Amount, 10) + " satoshis with Lightning Network")
  dw.Annotation(25, 365, "but all I got was a picture of his dog")
  mw.DrawImage(dw)
  mw.WriteImage("/var/www/lnd/tips/" + rHash + ".jpg")
  newWidth := uint(250)
  newHeight := uint(mw.GetImageHeight()*newWidth/mw.GetImageWidth())
  imerr = mw.ResizeImage(newWidth, newHeight, imagick.FILTER_LANCZOS, 1)
  if imerr != nil {
    panic(imerr)
  }
  mw.WriteImage("/var/www/lnd/tips/" + rHash + ".jpg_small.jpg")

	return GetImagePath(rHash)
}

