package module

import (
	"os"
	"strings"
)

func RegX() *GoBE {
	var printBannerV = os.Getenv("GOBE_PRINT_BANNER")
	if printBannerV == "" {
		printBannerV = "true"
	}

	return &GoBE{
		printBanner: strings.ToLower(printBannerV) == "true",
	}
}
