// +build !imagick
// Stubs to make code compile and behave same if it has image supportr or not.

package image

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
