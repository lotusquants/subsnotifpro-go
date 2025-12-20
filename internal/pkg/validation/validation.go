package validation

import (
	"fmt"
	"strings"

	"subsnotifpro-go/internal/pkg/apperrors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	// Register custom validators
	validate.RegisterValidation("package_name", validatePackageName)
	validate.RegisterValidation("product_id", validateProductID)
	validate.RegisterValidation("purchase_token", validatePurchaseToken)
}

// Common query parameter structures
type PlaystoreQueryParams struct {
	PackageName string `form:"package_name" validate:"required,package_name"`
	ProductID   string `form:"product_id" validate:"omitempty,product_id"`
	BasePlanID  string `form:"base_plan_id" validate:"omitempty,min=1,max=255"`
	OfferID     string `form:"offer_id" validate:"omitempty,min=1,max=255"`
	RegionCode  string `form:"region_code" validate:"omitempty,len=2"`
}

type AppstoreQueryParams struct {
	BundleID      string `form:"bundle_id" validate:"required,min=1,max=255"`
	TransactionID string `form:"transaction_id" validate:"omitempty,min=1,max=255"`
	ProductID     string `form:"product_id" validate:"omitempty,min=1,max=255"`
}

type SubscriptionQueryParams struct {
	PurchaseToken string `form:"purchase_token" validate:"required,purchase_token"`
	PackageName   string `form:"package_name" validate:"required,package_name"`
}

// Middleware to validate query parameters
func ValidateQuery(validatorStruct interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := c.ShouldBindQuery(validatorStruct); err != nil {
			// Parse validation errors
			var validationErrors []string

			if errs, ok := err.(validator.ValidationErrors); ok {
				for _, e := range errs {
					validationErrors = append(validationErrors, formatValidationError(e))
				}
			} else {
				validationErrors = append(validationErrors, err.Error())
			}

			appErr := apperrors.ValidationError(
				"Invalid query parameters",
				strings.Join(validationErrors, "; "),
			)

			apperrors.HandleAppError(c, appErr)
			c.Abort()
			return
		}

		// Store validated params in context
		c.Set("validated_params", validatorStruct)
		c.Next()
	}
}

// Middleware to validate JSON body
func ValidateJSON(validatorStruct interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := c.ShouldBindJSON(validatorStruct); err != nil {
			var validationErrors []string

			if errs, ok := err.(validator.ValidationErrors); ok {
				for _, e := range errs {
					validationErrors = append(validationErrors, formatValidationError(e))
				}
			} else {
				validationErrors = append(validationErrors, err.Error())
			}

			appErr := apperrors.ValidationError(
				"Invalid request body",
				strings.Join(validationErrors, "; "),
			)

			apperrors.HandleAppError(c, appErr)
			c.Abort()
			return
		}

		// Store validated body in context
		c.Set("validated_body", validatorStruct)
		c.Next()
	}
}

// Helper to get validated params from context
func GetValidatedParams(c *gin.Context) interface{} {
	if params, exists := c.Get("validated_params"); exists {
		return params
	}
	return nil
}

// Helper to get validated body from context
func GetValidatedBody(c *gin.Context) interface{} {
	if body, exists := c.Get("validated_body"); exists {
		return body
	}
	return nil
}

// Custom validators
func validatePackageName(fl validator.FieldLevel) bool {
	packageName := fl.Field().String()
	if len(packageName) < 3 || len(packageName) > 255 {
		return false
	}

	// Check if it looks like a valid package name (contains at least one dot)
	return strings.Contains(packageName, ".")
}

func validateProductID(fl validator.FieldLevel) bool {
	productID := fl.Field().String()
	if len(productID) < 1 || len(productID) > 255 {
		return false
	}

	// No special validation rules for product ID beyond length
	return true
}

func validatePurchaseToken(fl validator.FieldLevel) bool {
	token := fl.Field().String()
	if len(token) < 5 || len(token) > 1000 {
		return false
	}

	// Check if token contains only valid characters
	for _, char := range token {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '_' || char == '.') {
			return false
		}
	}

	return true
}

// Format validation errors for user-friendly messages
func formatValidationError(err validator.FieldError) string {
	field := err.Field()
	tag := err.Tag()

	switch tag {
	case "required":
		return fmt.Sprintf("Field '%s' is required", field)
	case "min":
		return fmt.Sprintf("Field '%s' must be at least %s characters", field, err.Param())
	case "max":
		return fmt.Sprintf("Field '%s' must be at most %s characters", field, err.Param())
	case "len":
		return fmt.Sprintf("Field '%s' must be exactly %s characters", field, err.Param())
	case "package_name":
		return fmt.Sprintf("Field '%s' must be a valid package name (e.g., com.example.app)", field)
	case "product_id":
		return fmt.Sprintf("Field '%s' must be a valid product ID", field)
	case "purchase_token":
		return fmt.Sprintf("Field '%s' must be a valid purchase token", field)
	default:
		return fmt.Sprintf("Field '%s' failed validation rule '%s'", field, tag)
	}
}

// Sanitization functions
func SanitizeStringInput(input string) string {
	// Trim whitespace and normalize
	return strings.TrimSpace(input)
}

func SanitizePlaystoreParams(params *PlaystoreQueryParams) {
	params.PackageName = SanitizeStringInput(params.PackageName)
	params.ProductID = SanitizeStringInput(params.ProductID)
	params.BasePlanID = SanitizeStringInput(params.BasePlanID)
	params.OfferID = SanitizeStringInput(params.OfferID)
	params.RegionCode = strings.ToUpper(SanitizeStringInput(params.RegionCode))
}

func SanitizeAppstoreParams(params *AppstoreQueryParams) {
	params.BundleID = SanitizeStringInput(params.BundleID)
	params.TransactionID = SanitizeStringInput(params.TransactionID)
	params.ProductID = SanitizeStringInput(params.ProductID)
}

func SanitizeSubscriptionParams(params *SubscriptionQueryParams) {
	params.PurchaseToken = SanitizeStringInput(params.PurchaseToken)
	params.PackageName = SanitizeStringInput(params.PackageName)
}

// Validation helper functions for manual validation
func ValidateRequiredFields(fields map[string]string) *apperrors.AppError {
	var missingFields []string

	for fieldName, fieldValue := range fields {
		if strings.TrimSpace(fieldValue) == "" {
			missingFields = append(missingFields, fieldName)
		}
	}

	if len(missingFields) > 0 {
		return apperrors.MultipleFieldErrors(missingFields)
	}

	return nil
}

func ValidatePackageNameManual(packageName string) *apperrors.AppError {
	if len(packageName) < 3 || len(packageName) > 255 || !strings.Contains(packageName, ".") {
		return apperrors.InvalidFieldError("package_name", "must be a valid package name (e.g., com.example.app)")
	}
	return nil
}

func ValidatePurchaseTokenManual(token string) *apperrors.AppError {
	if len(token) < 5 || len(token) > 1000 {
		return apperrors.InvalidFieldError("purchase_token", "must be between 5 and 1000 characters")
	}

	// Check if token contains only valid characters
	for _, char := range token {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '_' || char == '.') {
			return apperrors.InvalidFieldError("purchase_token", "contains invalid characters")
		}
	}

	return nil
}
