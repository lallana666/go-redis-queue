package filters

import (
	"net/http"
)

// WithCORS is a simple CORS implementation that wraps an http Handler.
func WithCORS(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		origin := req.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Requested-With, If-Modified-Since")
			w.Header().Set("Access-Control-Expose-Headers", "Date")

			// Stop here if its a preflight OPTIONS request
			if req.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}

		// Dispatch to the next handler
		handler.ServeHTTP(w, req)
	})
}
