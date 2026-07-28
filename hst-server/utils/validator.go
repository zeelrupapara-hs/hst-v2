package utils

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

// ValidatorMessage turns a validator error into one readable sentence.
func ValidatorMessage(err error) error {
	var verrs validator.ValidationErrors
	if !errors.As(err, &verrs) || len(verrs) == 0 {
		return err
	}

	field := verrs[0].Field()
	param := verrs[0].Param()

	switch verrs[0].Tag() {
	case "required":
		return fmt.Errorf("%s is required", field)
	case "email":
		return fmt.Errorf("%s is not a valid email", field)
	case "min":
		return fmt.Errorf("%s must be at least %s characters", field, param)
	case "max":
		return fmt.Errorf("%s must be at most %s characters", field, param)
	case "gte":
		return fmt.Errorf("%s must be greater than or equal to %s", field, param)
	case "lte":
		return fmt.Errorf("%s must be less than or equal to %s", field, param)
	case "oneof":
		return fmt.Errorf("%s must be one of %s", field, param)
	default:
		return fmt.Errorf("field %s is not valid", field)
	}
}
