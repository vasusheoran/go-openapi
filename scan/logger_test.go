package scan

import "testing"

func TestLogger_Info(t *testing.T) {
	logger := NewLogger(LogLevelInfo)
	logger.Debug("%s", "this should be skipped")
	logger.Info("%s", "this should be printed")
	logger.Warn("%s", "this should be printed")
	logger.Error("%s", "this should be printed")
	logger.Fatal("%s", "this should be printed")
}
