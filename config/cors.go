package config

import (
	"log"
	"net/url"
	"os"
	"strings"
)

func AllowedOrigins() []string {
	raw := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if raw == "" {
		if os.Getenv("ENV") != "production" {
			return []string{
				"http://localhost:5173",
				"http://localhost:3000",
				"http://127.0.0.1:5173",
				"http://127.0.0.1:3000",
			}
		}
		return nil
	}

	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		if o := strings.TrimSpace(part); o != "" {
			origins = append(origins, o)
		}
	}
	return origins
}

func normalizeOrigin(origin string) string {
	origin = strings.TrimSpace(origin)
	if origin == "" {
		return ""
	}

	u, err := url.Parse(origin)
	if err != nil {
		return strings.TrimRight(strings.ToLower(origin), "/")
	}

	scheme := strings.ToLower(u.Scheme)
	host := strings.ToLower(u.Hostname())
	port := u.Port()

	switch {
	case port == "" && scheme == "https":
		port = "443"
	case port == "" && scheme == "http":
		port = "80"
	}

	if (scheme == "http" && port == "80") || (scheme == "https" && port == "443") {
		return scheme + "://" + host
	}
	return scheme + "://" + host + ":" + port
}

func IsOriginAllowed(origin string) bool {
	origin = strings.TrimSpace(origin)
	if origin == "" {
		return true
	}

	normalized := normalizeOrigin(origin)
	for _, allowed := range AllowedOrigins() {
		if strings.EqualFold(normalizeOrigin(allowed), normalized) {
			return true
		}
	}
	return false
}

func LogOriginRejected(origin string) {
	allowed := AllowedOrigins()
	if len(allowed) == 0 {
		log.Printf("[CORS] Origin rejected: %q (CORS_ALLOWED_ORIGINS is empty in production)", origin)
		return
	}
	log.Printf("[CORS] Origin rejected: %q (allowed: %v)", origin, allowed)
}
