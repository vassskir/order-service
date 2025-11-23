package validation

import (
	"fmt"
	"regexp"
	"time"

	"L0/internal/models"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	validate.RegisterValidation("phone", validatePhone)
	validate.RegisterValidation("track_number", validateTrackNumber)
}

func ValidateOrder(order *models.Order) error {
	if err := validate.Struct(order); err != nil {
		return fmt.Errorf("order validation failed: %w", err)
	}
	if err := validateOrderRelations(order); err != nil {
		return err
	}

	return nil
}

func validatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()

	phoneRegex := `^\+?[1-9]\d{1,14}$`
	matched, _ := regexp.MatchString(phoneRegex, phone)
	return matched
}

func validateTrackNumber(fl validator.FieldLevel) bool {
	trackNumber := fl.Field().String()

	if len(trackNumber) == 0 || len(trackNumber) > 50 {
		return false
	}

	validChars := regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`)
	return validChars.MatchString(trackNumber)
}

func validateOrderRelations(order *models.Order) error {
	if order.Payment.Transaction != order.OrderUID {
		return fmt.Errorf("payment transaction must match order UID")
	}

	for _, item := range order.Items {
		if item.TrackNumber != order.TrackNumber {
			return fmt.Errorf("item track number must match order track number")
		}
	}

	if order.DateCreated.After(time.Now().Add(24 * time.Hour)) {
		return fmt.Errorf("order date cannot be in the future")
	}

	if order.DateCreated.Before(time.Now().Add(-365 * 24 * time.Hour)) {
		return fmt.Errorf("order date is too far in the past")
	}

	return nil
}

func IsValidationError(err error) bool {
	_, ok := err.(validator.ValidationErrors)
	return ok
}
