package middleware

import (
	"net/http"
	"strings"
	"time"
)

// SecurityConfig holds security middleware configuration
type SecurityConfig struct {
	// CORS settings
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
	MaxAge         int

	// Security headers
	ContentSecurityPolicy string
	XFrameOptions         string
	ReferrerPolicy        string
	PermissionsPolicy     string

	// Rate limiting
	RateLimitEnabled  bool
	RequestsPerMinute int
	BurstSize         int
}

// DefaultSecurityConfig returns secure default configuration
func DefaultSecurityConfig() SecurityConfig {
	return SecurityConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{
			"Accept",
			"Accept-Language",
			"Content-Type",
			"Content-Language",
			"Origin",
			"Authorization",
			"X-Requested-With",
			"X-Request-ID",
		},
		MaxAge:                86400, // 24 hours
		ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self'; media-src 'self'; object-src 'none'; child-src 'none'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'",
		XFrameOptions:         "DENY",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		PermissionsPolicy:     "geolocation=(), microphone=(), camera=(), payment=(), usb=(), magnetometer=(), gyroscope=(), accelerometer=()",
		RateLimitEnabled:      true,
		RequestsPerMinute:     100,
		BurstSize:             20,
	}
}

// SecurityMiddleware provides comprehensive security headers and CORS handling
type SecurityMiddleware struct {
	config SecurityConfig
}

// NewSecurityMiddleware creates a new security middleware with the given configuration
func NewSecurityMiddleware(config SecurityConfig) *SecurityMiddleware {
	return &SecurityMiddleware{
		config: config,
	}
}

// Handler returns the security middleware handler
func (s *SecurityMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add security headers
		s.addSecurityHeaders(w, r)

		// Handle CORS
		if s.handleCORS(w, r) {
			return
		}

		// Validate request
		if !s.validateRequest(w, r) {
			return
		}

		next.ServeHTTP(w, r)
	})
}

// addSecurityHeaders adds security-related HTTP headers
func (s *SecurityMiddleware) addSecurityHeaders(w http.ResponseWriter, r *http.Request) {
	// Basic security headers
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", s.config.XFrameOptions)
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	w.Header().Set("Referrer-Policy", s.config.ReferrerPolicy)

	// Content Security Policy
	if s.config.ContentSecurityPolicy != "" {
		w.Header().Set("Content-Security-Policy", s.config.ContentSecurityPolicy)
	}

	// Permissions Policy
	if s.config.PermissionsPolicy != "" {
		w.Header().Set("Permissions-Policy", s.config.PermissionsPolicy)
	}

	// Remove server information
	w.Header().Set("Server", "")

	// Add cache control for sensitive endpoints
	if s.isSensitiveEndpoint(r) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate, private")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
	}
}

// handleCORS handles Cross-Origin Resource Sharing
func (s *SecurityMiddleware) handleCORS(w http.ResponseWriter, r *http.Request) bool {
	origin := r.Header.Get("Origin")

	// Check if origin is allowed
	if origin != "" && s.isOriginAllowed(origin) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	} else if len(s.config.AllowedOrigins) == 1 && s.config.AllowedOrigins[0] == "*" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}

	// Handle preflight requests
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(s.config.AllowedMethods, ", "))
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(s.config.AllowedHeaders, ", "))
		w.Header().Set("Access-Control-Max-Age", "86400")
		w.WriteHeader(http.StatusNoContent)
		return true
	}

	return false
}

// validateRequest performs basic request validation
func (s *SecurityMiddleware) validateRequest(w http.ResponseWriter, r *http.Request) bool {
	// Check for suspicious user agents
	userAgent := r.Header.Get("User-Agent")
	if s.isSuspiciousUserAgent(userAgent) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return false
	}

	// Check for suspicious headers
	if s.hasSuspiciousHeaders(r) {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return false
	}

	// Check request size
	if r.ContentLength > 10*1024*1024 { // 10MB limit
		http.Error(w, "Request Entity Too Large", http.StatusRequestEntityTooLarge)
		return false
	}

	return true
}

// isOriginAllowed checks if the origin is in the allowed list
func (s *SecurityMiddleware) isOriginAllowed(origin string) bool {
	for _, allowedOrigin := range s.config.AllowedOrigins {
		if allowedOrigin == "*" || allowedOrigin == origin {
			return true
		}
		// Support wildcard subdomains
		if strings.HasPrefix(allowedOrigin, "*.") {
			domain := strings.TrimPrefix(allowedOrigin, "*.")
			if strings.HasSuffix(origin, domain) {
				return true
			}
		}
	}
	return false
}

// isSensitiveEndpoint checks if the endpoint contains sensitive data
func (s *SecurityMiddleware) isSensitiveEndpoint(r *http.Request) bool {
	sensitivePaths := []string{
		"/api/auth/",
		"/api/users/",
		"/api/admin/",
		"/api/settings/",
		"/api/webhooks/",
	}

	for _, path := range sensitivePaths {
		if strings.HasPrefix(r.URL.Path, path) {
			return true
		}
	}
	return false
}

// isSuspiciousUserAgent checks for suspicious user agents
func (s *SecurityMiddleware) isSuspiciousUserAgent(userAgent string) bool {
	if userAgent == "" {
		return true // Reject empty user agents
	}

	suspicious := []string{
		"sqlmap", "nikto", "masscan", "nmap", "whatweb", "dirb", "dirbuster",
		"gobuster", "wfuzz", "burp", "zap", "acunetix", "nessus", "openvas",
		"<script", "javascript:", "data:", "vbscript:",
	}

	lowerUA := strings.ToLower(userAgent)
	for _, pattern := range suspicious {
		if strings.Contains(lowerUA, pattern) {
			return true
		}
	}

	return false
}

// hasSuspiciousHeaders checks for suspicious headers
func (s *SecurityMiddleware) hasSuspiciousHeaders(r *http.Request) bool {
	// Check for common attack headers
	suspiciousHeaders := []string{
		"X-Forwarded-For",
		"X-Real-IP",
		"X-Client-IP",
		"X-Remote-IP",
		"X-Remote-Addr",
		"X-Cluster-Client-IP",
	}

	for _, header := range suspiciousHeaders {
		if value := r.Header.Get(header); value != "" {
			// Check for suspicious IP formats or injection attempts
			if strings.Contains(value, "<") || strings.Contains(value, ">") ||
				strings.Contains(value, "script") || strings.Contains(value, "javascript:") {
				return true
			}
		}
	}

	return false
}

// Rate limiting implementation (simplified)
type RateLimiter struct {
	requests map[string][]time.Time
	config   SecurityConfig
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config SecurityConfig) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		config:   config,
	}
}

// Allow checks if a request should be allowed
func (rl *RateLimiter) Allow(clientIP string) bool {
	if !rl.config.RateLimitEnabled {
		return true
	}

	now := time.Now()
	windowStart := now.Add(-time.Minute)

	// Clean old requests
	if times, exists := rl.requests[clientIP]; exists {
		var validTimes []time.Time
		for _, t := range times {
			if t.After(windowStart) {
				validTimes = append(validTimes, t)
			}
		}
		rl.requests[clientIP] = validTimes
	}

	// Check if rate limit exceeded
	if len(rl.requests[clientIP]) >= rl.config.RequestsPerMinute {
		return false
	}

	// Add current request
	rl.requests[clientIP] = append(rl.requests[clientIP], now)
	return true
}

// GetClientIP extracts client IP from request
func GetClientIP(r *http.Request) string {
	// Check forwarded headers (be careful with these in production)
	forwardedFor := r.Header.Get("X-Forwarded-For")
	if forwardedFor != "" {
		// Take the first IP in the list
		ips := strings.Split(forwardedFor, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check other headers
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	// Fall back to remote address
	ip := r.RemoteAddr
	if strings.Contains(ip, ":") {
		ip = strings.Split(ip, ":")[0]
	}
	return ip
}
