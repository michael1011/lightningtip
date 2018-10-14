package image

// Image type for holding configuration
type Image struct {
	ImageDir       string  `long:"imagedir" description:"Directory where generated images are placed"`
	ImageURLDir    string  `long:"imageurldir" description:"The directory part of the URL where browsers will find the images"`
	ImageFile      string  `long:"imagefile" description:"Full path to the image that will be served"`
	ImageText      string  `long:"imagetext" description:"The text. For amount, put {Amount}"`
	ImageTextX     float64 `long:"imagetextx" description:"X position for start of first line"`
	ImageTextY     float64 `long:"imagetexty" description:"Y position for start of first line"`
	ImageTextColor string  `long:"imagetextcolor" description:"Text Color"`
	ImageTextFont  string  `long:"imagetextfont" description:"Text Font"`
	ImageTextSize  float64 `long:"imagetextsize" description:"Text size (heigth in pixels)"`
}
