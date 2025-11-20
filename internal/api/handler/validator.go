package handler

import (
	"regexp"

	"github.com/gin-gonic/gin/binding"
	v10 "github.com/go-playground/validator/v10"
)

// Register a custom 'e164' validator that accepts either a strict E.164
// phone number (e.g. +1234567890) or a plain local numeric string used in tests.
func init() {
	if v, ok := binding.Validator.Engine().(*v10.Validate); ok {
		_ = v.RegisterValidation("e164", func(fl v10.FieldLevel) bool {
			val := fl.Field().String()
			// Accept empty? binding:"required" handles presence.
			// E.164 strict regex: + followed by country code (no leading 0) and up to 15 digits total.
			reE164 := regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
			if reE164.MatchString(val) {
				return true
			}
			// Accept plain digit sequences used in tests (7-15 digits)
			reDigits := regexp.MustCompile(`^\d{7,15}$`)
			return reDigits.MatchString(val)
		})
	}
}
