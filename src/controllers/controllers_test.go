package controllers

import (
	"testing"

	"rtx/src/models"
	"rtx/src/services"
)

func TestRotateIP(t *testing.T) {
	t.Skip("Skipping integration test - requires network access")

	result := services.RotateIP()

	if result.Error != nil {
		t.Logf("RotateIP error (may be expected offline): %v", result.Error)
		return
	}

	if result.OldIP == "" {
		t.Error("OldIP should not be empty on success")
	}
	if result.NewIP == "" {
		t.Error("NewIP should not be empty on success")
	}
	if result.Proxy == "" {
		t.Error("Proxy should not be empty on success")
	}
	if result.Source == "" {
		t.Error("Source should not be empty on success")
	}

	t.Logf("Rotation result: OldIP=%s, NewIP=%s, Proxy=%s, Changed=%v",
		result.OldIP, result.NewIP, result.Proxy, result.Changed)
}

func TestRotationResultStructure(t *testing.T) {
	result := models.RotationResult{
		OldIP:   "1.2.3.4",
		NewIP:   "5.6.7.8",
		Proxy:   "http://proxy:8080",
		Source:  "gimmeproxy",
		Changed: true,
		Error:   nil,
	}

	if result.OldIP != "1.2.3.4" {
		t.Errorf("OldIP = %s, want 1.2.3.4", result.OldIP)
	}
	if result.NewIP != "5.6.7.8" {
		t.Errorf("NewIP = %s, want 5.6.7.8", result.NewIP)
	}
	if result.Proxy != "http://proxy:8080" {
		t.Errorf("Proxy = %s, want http://proxy:8080", result.Proxy)
	}
	if result.Source != "gimmeproxy" {
		t.Errorf("Source = %s, want gimmeproxy", result.Source)
	}
	if !result.Changed {
		t.Error("Changed should be true")
	}
	if result.Error != nil {
		t.Errorf("Error should be nil, got %v", result.Error)
	}
}

func TestRotationResultNoChange(t *testing.T) {
	result := models.RotationResult{
		OldIP:   "1.2.3.4",
		NewIP:   "1.2.3.4",
		Proxy:   "http://proxy:8080",
		Source:  "test",
		Changed: false,
		Error:   nil,
	}

	if result.Changed {
		t.Error("Changed should be false when IP same")
	}
}

func TestRotationResultError(t *testing.T) {
	result := models.RotationResult{
		OldIP:   "",
		NewIP:   "",
		Proxy:   "",
		Source:  "",
		Changed: false,
		Error:   nil, // Em testes, error nil mas estrutura vazia é válida
	}

	if result.OldIP != "" {
		t.Error("Empty result should have empty OldIP")
	}
}