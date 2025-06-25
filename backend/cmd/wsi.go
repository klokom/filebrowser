// in backend/cmd/wsi.go

package cmd

import (
	"net/http"
	"time"

	"github.com/gtsteffaniak/filebrowser/backend/common/settings"
	"github.com/gtsteffaniak/go-logger/logger"
)

// validateWSIIntegration checks if the configured slideserver is available.
func validateWSIIntegration() {
	if settings.Config.Integrations.WSI.URL != "" {
		url := settings.Config.Integrations.WSI.URL
		if settings.Config.Integrations.WSI.InternalUrl != "" {
			url = settings.Config.Integrations.WSI.InternalUrl
		}

		client := &http.Client{
			Timeout: 5 * time.Second,
		}
		// Assuming your slideserver has a health check endpoint at /health or similar.
		// If not, a simple GET to the base URL can suffice.
		resp, err := client.Get(url)
		if err != nil {
			logger.Warningf("Could not reach the WSI slideserver at %s: %v", url, err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			logger.Warningf("WSI slideserver at %s returned non-OK status: %s", url, resp.Status)
		}
		logger.Info("WSI Integration Enabled")
	}
}