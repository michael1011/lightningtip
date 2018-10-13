// +build !imagick
// Stubs to make code compile and behave same if it has image supportr or not.

package image

// Image type for holding configuration
type Image struct {
	ImageDir             string  `long:"imagedir" description:"Directory where generated images are placed"`
	ImageURLDir          string  `long:"imageurldir" description:"The directory part of the URL where browsers will find the images"`
	ImageFile            string  `long:"imagefile" description:"Full path to the image that will be served"`
	ImageTextBeforeAmt   string  `long:"imagetextbeforeamt" description:"The text on the first line before the amount"`
	ImageTextAfterAmt    string  `long:"imagetextafteramt" description:"The text on the first line after the amount"`
	ImageTextSecondLine  string  `long:"imagetextsecondline" description:"The text on the second line"`
	ImageTextFirstLineX  float64 `long:"imagetextfirstlinex" description:"X position for start of first line"`
	ImageTextFirstLineY  float64 `long:"imagetextfirstliney" description:"Y position for start of first line"`
	ImageTextSecondLineX float64 `long:"imagetextsecondlinex" description:"X position for start of second line"`
	ImageTextSecondLineY float64 `long:"imagetextsecondliney" description:"Y position for start of second line"`
	ImageTextColor       string  `long:"imagetextcolor" description:"Text Color"`
	ImageTextFont        string  `long:"imagetextfont" description:"Text Font"`
	ImageTextSize        float64 `long:"imagetextsize" description:"Text size (heigth in pixels)"`
}

// ImageCfg is the Configuration variable
var ImageCfg Image

// SetImageCfg - Set package configuration from outside package
func SetImageCfg(Cfg Image) {
	ImageCfg = Cfg
}

// GetImagePath - Get the web path for the picture based on the payment hash
func GetImagePath(rHash string) string {
	return ("")
}

// GenerateImage - Generate a picture from hash and amount
func GenerateImage(rHash string, Amount int64) string {
	return ("")
}
