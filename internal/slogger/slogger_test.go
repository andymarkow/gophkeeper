package slogger

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogger(t *testing.T) {
	t.Run("new logger", func(t *testing.T) {
		logger := NewLogger(
			WithLevel(slog.LevelInfo),
			WithFormat(LogFormatText),
			WithAddSource(true),
		)

		assert.IsType(t, &slog.Logger{}, logger)
	})
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		expected slog.Level
		wantErr  bool
	}{
		{"valid debug", "debug", slog.LevelDebug, false},
		{"valid info", "info", slog.LevelInfo, false},
		{"valid warn", "warn", slog.LevelWarn, false},
		{"valid error", "error", slog.LevelError, false},
		{"uppercase level", "INFO", slog.LevelInfo, false},
		{"camelcase level", "InFo", slog.LevelInfo, false},
		{"invalid level", "unknown", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level, err := ParseLogLevel(tt.level)
			if tt.wantErr {
				require.Error(t, err)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, level)
		})
	}
}
