package http

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	// The "strings" import has been removed from here

	"github.com/gtsteffaniak/filebrowser/backend/common/settings"
)

// wsiProxyHandler is a standalone function that fits the middleware pattern.
func wsiProxyHandler(w http.ResponseWriter, r *http.Request, d *requestContext) (int, error) {
	// Use the internal URL for secure, container-to-container communication
	rawURL := settings.Config.Integrations.WSI.InternalUrl
	if rawURL == "" {
		// Fallback to the public URL if internal is not set
		rawURL = settings.Config.Integrations.WSI.URL
	}

	if rawURL == "" {
		// Return a real error, which the middleware will handle.
		return http.StatusInternalServerError, fmt.Errorf("WSI integration URL is not configured")
	}

	targetURL, err := url.Parse(rawURL)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("Invalid WSI server URL: %v", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// We no longer modify the r.URL.Path here. It is forwarded as-is.

	proxy.ServeHTTP(w, r)
	
	// Signal that the handler has successfully taken over the response.
	return 0, nil
}