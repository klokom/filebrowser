// in backend/preview/wsi.go

package preview

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"github.com/gtsteffaniak/filebrowser/backend/common/settings"
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
func GetWSIMetadata(filePath string) (map[string]interface{}, error) {
	internalURL := settings.Config.Integrations.WSI.InternalUrl
	if internalURL == "" {
		return nil, fmt.Errorf("slideserver internal URL is not configured")
	}

	encodedPath := url.PathEscape(strings.TrimPrefix(filePath, "/"))
	fullURL := fmt.Sprintf("%s/wsi/%s.metadata", internalURL, encodedPath)

	resp, err := http.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("failed to request metadata from slideserver: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("slideserver returned non-200 status: %s", resp.Status)
	}

	var metadata map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return nil, fmt.Errorf("failed to decode metadata JSON: %w", err)
	}

	return metadata, nil
}