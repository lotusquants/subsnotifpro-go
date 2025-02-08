// internal/google_playstore/logger/logger.go
package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Log is the global structured logger
var Log = logrus.New()

func init() {
	Log.SetFormatter(&logrus.JSONFormatter{}) // ✅ JSON format for logs
	Log.SetOutput(os.Stdout)                  // ✅ Log to console
	Log.SetLevel(logrus.InfoLevel)            // ✅ Default level: INFO
}
