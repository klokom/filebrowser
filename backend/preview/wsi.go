// in backend/preview/wsi.go

package preview

import (
	"strings"
)

// SupportedWSITypes is a map of file extensions for Whole Slide Images.
var SupportedWSITypes = map[string]bool{
	".svs":    true,
	".ndpi":   true,
	".ome.tif":true,
	".ome.tiff":true,
	".vmsi":   true,
	".vmu":    true,
	".scn":    true,
	".mrxs":   true,
	".bif":    true,
}

// IsWSI checks if the file is a supported WSI type by its name.
func IsWSI(name string) bool {
	lowerName := strings.ToLower(name)
	for ext := range SupportedWSITypes {
		if strings.HasSuffix(lowerName, ext) {
			return true
		}
	}
	return false
}