package validator

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"subsnotifpro-go/internal/pkg/logger"

	"github.com/gin-gonic/gin"
)

// Google Cloud Pub/Sub IP ranges (these should be updated regularly)
var googleCloudPubSubIPRanges = []string{
	"64.233.160.0/19",
	"66.102.0.0/20",
	"66.249.80.0/20",
	"72.14.192.0/18",
	"74.125.0.0/16",
	"108.177.8.0/21",
	"173.194.0.0/16",
	"209.85.128.0/17",
	"216.58.192.0/19",
	"216.239.32.0/19",
	// Add more IP ranges as needed
}

// Apple App Store Server Notification IP ranges (these should be updated regularly)
var appleAppStoreIPRanges = []string{
	"17.0.0.0/8",
	"2620:149::/32",
	"2403:300::/32",
	// Add more IP ranges as needed
}

// ValidateGoogleCloudPubSubIP validates that the request comes from Google Cloud Pub/Sub
func ValidateGoogleCloudPubSubIP(c *gin.Context) bool {
	clientIP := getClientIP(c)
	logger.Log.Infof("Validating Google Cloud Pub/Sub IP: %s", clientIP)

	for _, ipRange := range googleCloudPubSubIPRanges {
		if isIPInRange(clientIP, ipRange) {
			logger.Log.Infof("✅ Valid Google Cloud Pub/Sub IP: %s", clientIP)
			return true
		}
	}

	logger.Log.Warnf("❌ Invalid Google Cloud Pub/Sub IP: %s", clientIP)
	return false
}

// ValidateAppleAppStoreIP validates that the request comes from Apple App Store servers
func ValidateAppleAppStoreIP(c *gin.Context) bool {
	clientIP := getClientIP(c)
	logger.Log.Infof("Validating Apple App Store IP: %s", clientIP)

	for _, ipRange := range appleAppStoreIPRanges {
		if isIPInRange(clientIP, ipRange) {
			logger.Log.Infof("✅ Valid Apple App Store IP: %s", clientIP)
			return true
		}
	}

	logger.Log.Warnf("❌ Invalid Apple App Store IP: %s", clientIP)
	return false
}

// getClientIP extracts the client IP from the request
func getClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header first
	forwarded := c.GetHeader("X-Forwarded-For")
	if forwarded != "" {
		// Take the first IP in the list
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	realIP := c.GetHeader("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to remote address
	return c.ClientIP()
}

// isIPInRange checks if an IP address is within a CIDR range
func isIPInRange(ipStr, cidrRange string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	_, network, err := net.ParseCIDR(cidrRange)
	if err != nil {
		logger.Log.Errorf("Invalid CIDR range: %s, error: %v", cidrRange, err)
		return false
	}

	return network.Contains(ip)
}

// ValidateGoogleCloudPubSubRequest validates the complete Google Cloud Pub/Sub request
func ValidateGoogleCloudPubSubRequest(c *gin.Context) error {
	// Skip IP validation in development mode
	if gin.Mode() == gin.DebugMode {
		logger.Log.Info("Skipping IP validation in development mode")
		return validateBasicRequest(c)
	}

	// 1. Validate IP address
	if !ValidateGoogleCloudPubSubIP(c) {
		return fmt.Errorf("invalid IP address for Google Cloud Pub/Sub")
	}

	return validateBasicRequest(c)
}

// ValidateAppleAppStoreRequest validates the complete Apple App Store request
func ValidateAppleAppStoreRequest(c *gin.Context) error {
	// Skip IP validation in development mode
	if gin.Mode() == gin.DebugMode {
		logger.Log.Info("Skipping IP validation in development mode")
		return validateBasicRequest(c)
	}

	// 1. Validate IP address
	if !ValidateAppleAppStoreIP(c) {
		return fmt.Errorf("invalid IP address for Apple App Store")
	}

	return validateBasicRequest(c)
}

// validateBasicRequest validates common request properties
func validateBasicRequest(c *gin.Context) error {
	// 1. Validate HTTP method
	if c.Request.Method != http.MethodPost {
		return fmt.Errorf("invalid HTTP method: %s, expected POST", c.Request.Method)
	}

	// 2. Validate Content-Type
	contentType := c.GetHeader("Content-Type")
	if contentType != "application/json" {
		return fmt.Errorf("invalid Content-Type: %s, expected application/json", contentType)
	}

	return nil
}
