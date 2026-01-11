package output

import (
	"bytes"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name  string
		level Level
	}{
		{"Quiet level", LevelQuiet},
		{"Normal level", LevelNormal},
		{"Verbose level", LevelVerbose},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := New(tt.level)
			if log.Level() != tt.level {
				t.Errorf("New(%v).Level() = %v, want %v", tt.level, log.Level(), tt.level)
			}
		})
	}
}

func TestDebug(t *testing.T) {
	tests := []struct {
		name       string
		level      Level
		wantOutput bool
	}{
		{"Quiet suppresses debug", LevelQuiet, false},
		{"Normal suppresses debug", LevelNormal, false},
		{"Verbose shows debug", LevelVerbose, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			log := New(tt.level)
			log.out = &buf

			log.Debug("test message")

			hasOutput := buf.Len() > 0
			if hasOutput != tt.wantOutput {
				t.Errorf("Debug() output = %v, want output = %v", hasOutput, tt.wantOutput)
			}
		})
	}
}

func TestInfo(t *testing.T) {
	tests := []struct {
		name       string
		level      Level
		wantOutput bool
	}{
		{"Quiet suppresses info", LevelQuiet, false},
		{"Normal shows info", LevelNormal, true},
		{"Verbose shows info", LevelVerbose, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			log := New(tt.level)
			log.out = &buf

			log.Info("test message")

			hasOutput := buf.Len() > 0
			if hasOutput != tt.wantOutput {
				t.Errorf("Info() output = %v, want output = %v", hasOutput, tt.wantOutput)
			}
		})
	}
}

func TestResult(t *testing.T) {
	tests := []struct {
		name  string
		level Level
	}{
		{"Quiet shows result", LevelQuiet},
		{"Normal shows result", LevelNormal},
		{"Verbose shows result", LevelVerbose},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			log := New(tt.level)
			log.out = &buf

			log.Result("test message")

			if buf.Len() == 0 {
				t.Errorf("Result() should always output, got empty buffer")
			}
		})
	}
}

func TestError(t *testing.T) {
	tests := []struct {
		name  string
		level Level
	}{
		{"Quiet shows error", LevelQuiet},
		{"Normal shows error", LevelNormal},
		{"Verbose shows error", LevelVerbose},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			log := New(tt.level)
			log.errOut = &buf

			log.Error("test error")

			if buf.Len() == 0 {
				t.Errorf("Error() should always output to stderr, got empty buffer")
			}
		})
	}
}

func TestFormatting(t *testing.T) {
	var buf bytes.Buffer
	log := New(LevelVerbose)
	log.out = &buf

	log.Info("value: %d, string: %s", 42, "hello")

	expected := "value: 42, string: hello"
	if buf.String() != expected {
		t.Errorf("Info() = %q, want %q", buf.String(), expected)
	}
}
