package controllers

import (
	"rtx/src/models"
	"rtx/src/services"
)

func RotateIP() models.RotationResult {
	return services.RotateIP()
}
