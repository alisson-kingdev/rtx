// Package controllers keeps the CLI-facing rotation entry point.
package controllers

import (
	"context"

	"rtx/src/models"
	"rtx/src/services"
)

// RotateIP runs an IP rotation with the given limits.
func RotateIP(ctx context.Context, config services.RotationConfig) (models.RotationResult, error) {
	return services.RotateIP(ctx, config)
}
