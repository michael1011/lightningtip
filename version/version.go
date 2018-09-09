package version

import "fmt"

// Version is the version of LightningTip
const Version = "1.1.0-dev"

// PrintVersion prints the version of LightningTip to the console
func PrintVersion() {
	fmt.Println("LightningTip version " + Version)
}
