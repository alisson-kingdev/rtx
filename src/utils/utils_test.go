package utils

import (
	"strings"
	"testing"
)

func TestGetVersion(t *testing.T) {
	version := GetVersion()
	if version == "" {
		t.Error("GetVersion() should return non-empty string")
	}
	if version == "(devel)" {
		t.Error("GetVersion() should not return (devel)")
	}
	t.Logf("Version: %s", version)
}

func TestBanner(t *testing.T) {
	banner := Banner()
	if banner == "" {
		t.Error("Banner() should return non-empty string")
	}
	if !strings.Contains(banner, "RTX") {
		t.Error("Banner() should contain 'RTX'")
	}
	if !strings.Contains(banner, "Author: MRX") {
		t.Error("Banner() should contain 'Author: MRX'")
	}
	version := GetVersion()
	if !strings.Contains(banner, version) {
		t.Errorf("Banner() should contain version '%s'", version)
	}
	t.Logf("Banner output:\n%s", banner)
}

func TestSection(t *testing.T) {
	result := Section("Test Title")
	expected := "\n[*] Test Title\n"
	if result != expected {
		t.Errorf("Section() = %q, want %q", result, expected)
	}
}

func TestSuccess(t *testing.T) {
	result := Success("Operation completed")
	if !strings.Contains(result, "[+]") {
		t.Error("Success() should contain '[+]' prefix")
	}
	if !strings.Contains(result, "Operation completed") {
		t.Error("Success() should contain the message")
	}
}

func TestError(t *testing.T) {
	result := Error("Something failed")
	if !strings.Contains(result, "[!]") {
		t.Error("Error() should contain '[!]' prefix")
	}
	if !strings.Contains(result, "Something failed") {
		t.Error("Error() should contain the message")
	}
}

func TestInfo(t *testing.T) {
	result := Info("Information message")
	if !strings.Contains(result, "[*]") {
		t.Error("Info() should contain '[*]' prefix")
	}
	if !strings.Contains(result, "Information message") {
		t.Error("Info() should contain the message")
	}
}

func TestWarning(t *testing.T) {
	result := Warning("Warning message")
	if !strings.Contains(result, "[*]") {
		t.Error("Warning() should contain '[*]' prefix")
	}
	if !strings.Contains(result, "Warning message") {
		t.Error("Warning() should contain the message")
	}
}

func TestPrintKV(t *testing.T) {
	result := PrintKV("TestKey", "TestValue")
	if result == "" {
		t.Error("PrintKV() should return non-empty string")
	}
	if !strings.Contains(result, "TestKey") {
		t.Error("PrintKV() should contain key")
	}
	if !strings.Contains(result, "TestValue") {
		t.Error("PrintKV() should contain value")
	}
}

func TestColorize(t *testing.T) {
	result := Colorize(CyanColor, "test text")
	if !strings.Contains(result, "test text") {
		t.Error("Colorize() should contain the original text")
	}
}

func TestLogger(t *testing.T) {
	logger := &Logger{}
	logger.Success("success message")
	logger.Error("error message")
	logger.Info("info message")
	logger.Warning("warning message")
	logger.Log("success", "test")
	logger.Log("error", "test")
	logger.Log("info", "test")
	logger.Log("warning", "test")
	logger.Log("unknown", "test")
	Log.Success("package success")
	Log.Error("package error")
	Log.Info("package info")
	Log.Warning("package warning")
}

func TestBannerStructure(t *testing.T) {
	banner := Banner()
	lines := strings.Split(banner, "\n")
	if len(lines) < 9 {
		t.Errorf("Banner should have at least 9 lines, got %d", len(lines))
	}
	foundVersion := false
	for _, line := range lines {
		if strings.Contains(line, "RTX - IP ROTATOR v") {
			foundVersion = true
			break
		}
	}
	if !foundVersion {
		t.Error("Banner should contain version line")
	}
	foundAuthor := false
	for _, line := range lines {
		if strings.Contains(line, "Author: MRX") {
			foundAuthor = true
			break
		}
	}
	if !foundAuthor {
		t.Error("Banner should contain author line")
	}
}

func TestConstants(t *testing.T) {
	if CyanColor == nil {
		t.Error("CyanColor should not be nil")
	}
	if MagentaColor == nil {
		t.Error("MagentaColor should not be nil")
	}
	if YellowColor == nil {
		t.Error("YellowColor should not be nil")
	}
	if GreenColor == nil {
		t.Error("GreenColor should not be nil")
	}
	if RedColor == nil {
		t.Error("RedColor should not be nil")
	}
	if BlueColor == nil {
		t.Error("BlueColor should not be nil")
	}
	if WhiteColor == nil {
		t.Error("WhiteColor should not be nil")
	}
}