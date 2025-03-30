package service

import "time"

func parseTimeOrNil(timeStr string) time.Time {
	if timeStr == "" {
		return time.Time{}
	}

	t, err := time.Parse(time.RFC3339Nano, timeStr)
	if err != nil {
		return time.Time{}
	}
	return t
}
