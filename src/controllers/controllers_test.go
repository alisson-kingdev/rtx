package controllers

import (
	"context"
	"errors"
	"testing"

	"rtx/src/models"
	"rtx/src/services"
)

func TestRotateIPPropagatesCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := RotateIP(ctx, services.RotationConfig{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}

func TestDefaultRotationConfigIsBounded(t *testing.T) {
	t.Parallel()

	config := services.DefaultRotationConfig()
	if config.Timeout <= 0 || config.MaxProxies <= 0 || config.Concurrency <= 0 {
		t.Fatalf("unbounded default rotation config: %+v", config)
	}
}

func TestRotationResultModelRoundTrip(t *testing.T) {
	t.Parallel()

	value := models.RotationResult{OldIP: "1.1.1.1", NewIP: "2.2.2.2", Changed: true}
	if value.OldIP != "1.1.1.1" || value.NewIP != "2.2.2.2" || !value.Changed {
		t.Fatalf("unexpected model value: %+v", value)
	}
}
