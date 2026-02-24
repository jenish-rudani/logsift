package logsift

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

// setupTest resets global logger state and returns a buffer capturing log output.
// Uses JSON format for reliable field assertions.
func setupTest(t *testing.T) *bytes.Buffer {
	t.Helper()
	buf := &bytes.Buffer{}
	SetOutput(buf)
	SetLevel("debug")
	SetFormat("json")
	SetSourceFormat("short")
	SetAllowEmptyFilter(false)
	UpdateFilter(make(map[string]bool))
	return buf
}

// parseLogEntry parses a single JSON log line from the buffer.
func parseLogEntry(t *testing.T, buf *bytes.Buffer) map[string]interface{} {
	t.Helper()
	var entry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse log output as JSON: %v\nraw: %s", err, buf.String())
	}
	return entry
}

// --- Configuration tests ---

func TestSetLevel(t *testing.T) {
	buf := setupTest(t)

	SetLevel("warn")
	Info("should not appear")
	if buf.Len() != 0 {
		t.Error("expected Info to be suppressed at warn level")
	}

	Warn("should appear")
	if buf.Len() == 0 {
		t.Error("expected Warn to produce output at warn level")
	}
}

func TestSetLevel_InvalidFallsBackToInfo(t *testing.T) {
	setupTest(t)

	SetLevel("notavalidlevel")
	if GetLevel() != "info" {
		t.Errorf("expected invalid level to fall back to 'info', got %q", GetLevel())
	}
}

func TestGetLevel(t *testing.T) {
	setupTest(t)

	SetLevel("debug")
	if got := GetLevel(); got != "debug" {
		t.Errorf("expected 'debug', got %q", got)
	}

	SetLevel("error")
	if got := GetLevel(); got != "error" {
		t.Errorf("expected 'error', got %q", got)
	}
}

func TestSetFormat_JSON(t *testing.T) {
	buf := setupTest(t)

	SetFormat("json")
	Info("test message")

	var entry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("expected valid JSON output, got error: %v", err)
	}
	if entry["msg"] != "test message" {
		t.Errorf("expected msg='test message', got %v", entry["msg"])
	}
}

func TestGetFormat(t *testing.T) {
	setupTest(t)

	SetFormat("json")
	if got := GetFormat(); got != "json" {
		t.Errorf("expected 'json', got %q", got)
	}

	SetFormat("nocolor")
	if got := GetFormat(); got != "nocolor" {
		t.Errorf("expected 'nocolor', got %q", got)
	}

	SetFormat("text")
	if got := GetFormat(); got != "text" {
		t.Errorf("expected 'text', got %q", got)
	}
}

func TestSetSourceFormat(t *testing.T) {
	setupTest(t)

	SetSourceFormat("long")
	if got := GetSourceFormat(); got != "long" {
		t.Errorf("expected 'long', got %q", got)
	}

	SetSourceFormat("short")
	if got := GetSourceFormat(); got != "short" {
		t.Errorf("expected 'short', got %q", got)
	}

	// Invalid defaults to short
	SetSourceFormat("invalid")
	if got := GetSourceFormat(); got != "short" {
		t.Errorf("expected invalid to default to 'short', got %q", got)
	}
}

func TestIsDebugEnabled(t *testing.T) {
	setupTest(t)

	SetLevel("debug")
	if !IsDebugEnabled() {
		t.Error("expected IsDebugEnabled() to be true at debug level")
	}

	SetLevel("info")
	if IsDebugEnabled() {
		t.Error("expected IsDebugEnabled() to be false at info level")
	}
}

// --- Source tracking tests ---

func TestWithSource_Short(t *testing.T) {
	buf := setupTest(t)
	SetSourceFormat("short")

	Info("source test")

	entry := parseLogEntry(t, buf)
	source, ok := entry["source"].(string)
	if !ok {
		t.Fatal("expected 'source' field in log entry")
	}
	// Should be just filename:line, no directory separator
	if strings.Contains(source, "/") {
		t.Errorf("short source format should not contain '/', got %q", source)
	}
	if !strings.Contains(source, "log_test.go:") {
		t.Errorf("expected source to contain 'log_test.go:', got %q", source)
	}
}

func TestWithSource_Long(t *testing.T) {
	buf := setupTest(t)
	SetSourceFormat("long")

	Info("source test")

	entry := parseLogEntry(t, buf)
	source, ok := entry["source"].(string)
	if !ok {
		t.Fatal("expected 'source' field in log entry")
	}
	// Long format should contain a directory separator
	if !strings.Contains(source, "/") {
		t.Errorf("long source format should contain '/', got %q", source)
	}
}

// --- Structured fields tests ---

func TestWith(t *testing.T) {
	buf := setupTest(t)

	logger := With("request_id", "abc-123")
	logger.Info("with test")

	entry := parseLogEntry(t, buf)
	if entry["request_id"] != "abc-123" {
		t.Errorf("expected request_id='abc-123', got %v", entry["request_id"])
	}
}

func TestWithFields(t *testing.T) {
	buf := setupTest(t)

	logger := WithFields(map[string]interface{}{
		"user_id":  42,
		"username": "alice",
	})
	logger.Info("fields test")

	entry := parseLogEntry(t, buf)
	if entry["username"] != "alice" {
		t.Errorf("expected username='alice', got %v", entry["username"])
	}
	// JSON numbers decode as float64
	if entry["user_id"] != float64(42) {
		t.Errorf("expected user_id=42, got %v", entry["user_id"])
	}
}

func TestWith_Chaining(t *testing.T) {
	buf := setupTest(t)

	logger := With("a", 1).With("b", 2)
	logger.Info("chained")

	entry := parseLogEntry(t, buf)
	if entry["a"] != float64(1) {
		t.Errorf("expected a=1, got %v", entry["a"])
	}
	if entry["b"] != float64(2) {
		t.Errorf("expected b=2, got %v", entry["b"])
	}
}

func TestWith_DoesNotMutateParent(t *testing.T) {
	buf := setupTest(t)

	parent := With("a", 1)
	_ = parent.With("b", 2) // child adds "b"

	parent.Info("parent only")

	entry := parseLogEntry(t, buf)
	if _, ok := entry["b"]; ok {
		t.Error("expected parent logger NOT to have child's 'b' field")
	}
	if entry["a"] != float64(1) {
		t.Errorf("expected parent to have a=1, got %v", entry["a"])
	}
}

func TestDefault(t *testing.T) {
	setupTest(t)

	logger := Default()
	if logger == nil {
		t.Fatal("expected Default() to return non-nil Logger")
	}
	// Verify it implements Logger by calling a method
	logger.Info("default logger works")
}

// --- Filtered logging integration tests ---

func TestDebugFilter_Allowed(t *testing.T) {
	buf := setupTest(t)

	AddFilter("auth")
	DebugFilter("auth", "secret message")

	if buf.Len() == 0 {
		t.Error("expected DebugFilter to produce output when filter is active")
	}
	if !strings.Contains(buf.String(), "secret message") {
		t.Errorf("expected output to contain 'secret message', got %s", buf.String())
	}
}

func TestDebugFilter_Blocked(t *testing.T) {
	buf := setupTest(t)

	// Do NOT add filter
	DebugFilter("auth", "secret message")

	if buf.Len() != 0 {
		t.Error("expected DebugFilter to suppress output when filter is not active")
	}
}

func TestInfoFilter_Allowed(t *testing.T) {
	buf := setupTest(t)

	AddFilter("db")
	InfoFilter("db", "query executed")

	if buf.Len() == 0 {
		t.Error("expected InfoFilter to produce output when filter is active")
	}
}

func TestInfoFilter_Blocked(t *testing.T) {
	buf := setupTest(t)

	InfoFilter("db", "query executed")

	if buf.Len() != 0 {
		t.Error("expected InfoFilter to suppress output when filter is not active")
	}
}

func TestDebugFilters_MatchAny(t *testing.T) {
	buf := setupTest(t)

	AddFilter("db")
	DebugFilters([]string{"auth", "db"}, "multi-filter")

	if buf.Len() == 0 {
		t.Error("expected DebugFilters to log when any filter matches")
	}
}

func TestDebugFilters_NoneMatch(t *testing.T) {
	buf := setupTest(t)

	AddFilter("cache")
	DebugFilters([]string{"auth", "db"}, "multi-filter")

	if buf.Len() != 0 {
		t.Error("expected DebugFilters to suppress when no filter matches")
	}
}

func TestDebugFilterf_Formatting(t *testing.T) {
	buf := setupTest(t)

	AddFilter("x")
	DebugFilterf("x", "count=%d", 42)

	if !strings.Contains(buf.String(), "count=42") {
		t.Errorf("expected formatted output 'count=42', got %s", buf.String())
	}
}

func TestFilteredLog_AfterRemove(t *testing.T) {
	buf := setupTest(t)

	AddFilter("x")
	RemoveFilter("x")
	DebugFilter("x", "should not appear")

	if buf.Len() != 0 {
		t.Error("expected filter to be inactive after RemoveFilter")
	}
}

func TestFilteredLog_AfterUpdateFilter(t *testing.T) {
	buf := setupTest(t)

	UpdateFilter(map[string]bool{"a": true})

	DebugFilter("a", "allowed")
	if buf.Len() == 0 {
		t.Error("expected 'a' to pass after UpdateFilter")
	}

	buf.Reset()
	DebugFilter("b", "blocked")
	if buf.Len() != 0 {
		t.Error("expected 'b' to be blocked after UpdateFilter with only 'a'")
	}
}

// --- HTTP Handler tests ---

func TestHandler_SetLevel(t *testing.T) {
	setupTest(t)
	SetLevel("info") // start at info

	handler := Handler()
	req := httptest.NewRequest("GET", "/log?level=debug", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := GetLevel(); got != "debug" {
		t.Errorf("expected level 'debug' after handler, got %q", got)
	}
}

func TestHandler_SetFormat(t *testing.T) {
	setupTest(t)

	handler := Handler()
	req := httptest.NewRequest("GET", "/log?format=nocolor", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := GetFormat(); got != "nocolor" {
		t.Errorf("expected format 'nocolor' after handler, got %q", got)
	}
}

func TestHandler_SetSourceFormat(t *testing.T) {
	setupTest(t)

	handler := Handler()
	req := httptest.NewRequest("GET", "/log?sourceFormat=long", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := GetSourceFormat(); got != "long" {
		t.Errorf("expected sourceFormat 'long' after handler, got %q", got)
	}
}

func TestHandler_SetFilter(t *testing.T) {
	buf := setupTest(t)

	handler := Handler()
	req := httptest.NewRequest("GET", "/log?filter=auth,db", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	buf.Reset()
	DebugFilter("auth", "should appear")
	if buf.Len() == 0 {
		t.Error("expected 'auth' filter to be active after handler set it")
	}

	buf.Reset()
	DebugFilter("db", "should appear")
	if buf.Len() == 0 {
		t.Error("expected 'db' filter to be active after handler set it")
	}

	buf.Reset()
	DebugFilter("cache", "should not appear")
	if buf.Len() != 0 {
		t.Error("expected 'cache' filter to be inactive")
	}
}

func TestHandler_ResetFilter(t *testing.T) {
	buf := setupTest(t)

	AddFilter("auth")
	handler := Handler()
	req := httptest.NewRequest("GET", "/log?resetFilter=true", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	buf.Reset()
	DebugFilter("auth", "should not appear")
	if buf.Len() != 0 {
		t.Error("expected filters to be cleared after resetFilter=true")
	}
}

func TestHandler_AllowEmptyFilter(t *testing.T) {
	buf := setupTest(t)

	handler := Handler()
	req := httptest.NewRequest("GET", "/log?allowEmptyFilter=true", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	buf.Reset()
	DebugFilter("nonexistent", "should appear with empty filter + allowEmpty")
	if buf.Len() == 0 {
		t.Error("expected filtered log to pass with allowEmptyFilter=true and no filters set")
	}
}

func TestHandler_InvalidBoolParams(t *testing.T) {
	setupTest(t)

	handler := Handler()

	// Invalid allowEmptyFilter
	req := httptest.NewRequest("GET", "/log?allowEmptyFilter=notabool", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	// Should not panic — just return

	// Invalid resetFilter
	req = httptest.NewRequest("GET", "/log?resetFilter=notabool", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	// Should not panic
}

func TestHandler_MultipleParams(t *testing.T) {
	setupTest(t)

	handler := Handler()
	req := httptest.NewRequest("GET", "/log?level=warn&format=nocolor&sourceFormat=long", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := GetLevel(); got != "warning" {
		t.Errorf("expected level 'warning', got %q", got)
	}
	if got := GetFormat(); got != "nocolor" {
		t.Errorf("expected format 'nocolor', got %q", got)
	}
	if got := GetSourceFormat(); got != "long" {
		t.Errorf("expected sourceFormat 'long', got %q", got)
	}
}

func TestHandler_NoParams(t *testing.T) {
	setupTest(t)
	originalLevel := GetLevel()

	handler := Handler()
	req := httptest.NewRequest("GET", "/log", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := GetLevel(); got != originalLevel {
		t.Errorf("expected level to remain %q with no params, got %q", originalLevel, got)
	}
}

// --- ParseFilters tests ---

func TestParseFilters_Basic(t *testing.T) {
	result := ParseFilters("auth,db,cache")
	expected := map[string]bool{"auth": true, "db": true, "cache": true}

	if len(result) != len(expected) {
		t.Fatalf("expected %d entries, got %d", len(expected), len(result))
	}
	for k := range expected {
		if !result[k] {
			t.Errorf("expected key %q to be true", k)
		}
	}
}

func TestParseFilters_EmptyString(t *testing.T) {
	result := ParseFilters("")
	if len(result) != 0 {
		t.Errorf("expected empty map for empty string, got %v", result)
	}
}

func TestParseFilters_TrailingComma(t *testing.T) {
	result := ParseFilters("auth,db,")
	if len(result) != 2 {
		t.Errorf("expected 2 entries (trailing comma ignored), got %d: %v", len(result), result)
	}
	if !result["auth"] || !result["db"] {
		t.Errorf("expected auth and db to be present, got %v", result)
	}
}

func TestParseFilters_SingleValue(t *testing.T) {
	result := ParseFilters("auth")
	if len(result) != 1 || !result["auth"] {
		t.Errorf("expected {auth: true}, got %v", result)
	}
}
