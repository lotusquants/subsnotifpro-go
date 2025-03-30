package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"subsnotifpro-go/internal/playstore/products/models"
	"time"
)

// ✅ Helper Function to Check if Money is Zero
func IsZeroMoney(m models.Money) bool {
	return m.CurrencyCode == "" && m.Units == 0 && m.Nanos == 0
}

// ✅ ParseISO8601Duration - Converts ISO 8601 duration (e.g., "P1Y2M3W4D") to time.Duration
func ParseISO8601Duration(durationStr string, startDate time.Time) (time.Duration, error) {
	// 🔹 ISO 8601 Duration Pattern (Capturing Groups for Years, Months, Weeks, and Days)
	re := regexp.MustCompile(`P(?:(\d+)Y)?(?:(\d+)M)?(?:(\d+)W)?(?:(\d+)D)?`)

	// 🔹 Extract Matches
	matches := re.FindStringSubmatch(durationStr)
	if matches == nil {
		return 0, fmt.Errorf("invalid ISO 8601 duration format: %s", durationStr)
	}

	// 🔹 Convert Extracted Strings to Integers (Default: 0 if not present)
	years, _ := strconv.Atoi(defaultZero(matches[1]))  // Years
	months, _ := strconv.Atoi(defaultZero(matches[2])) // Months
	weeks, _ := strconv.Atoi(defaultZero(matches[3]))  // Weeks
	days, _ := strconv.Atoi(defaultZero(matches[4]))   // Days

	// 🔹 Start Date (Ensure we add months correctly)
	finalDate := startDate.AddDate(years, months, days+(weeks*7)) // Uses Go's built-in time.Date calculation

	// 🔹 Calculate Accurate Duration
	duration := finalDate.Sub(startDate)

	return duration, nil
}

// ✅ Helper: Returns "0" if input is empty
func defaultZero(input string) string {
	if input == "" {
		return "0"
	}
	return input
}

// ✅ Utility: Subtract Money Safely (Handles Negative Nanos)
func SubtractMoney(a, b models.Money) models.Money {
	result := a

	// Perform safe subtraction for nanos
	result.Nanos -= b.Nanos
	if result.Nanos < 0 {
		result.Units -= 1
		result.Nanos += 1e9
	}

	// Subtract Units
	result.Units -= b.Units

	return result
}
