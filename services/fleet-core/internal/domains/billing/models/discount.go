package models

// DiscountType defines the type of discount
type DiscountType string

const (
	DiscountNone       DiscountType = "none"
	DiscountPercentage DiscountType = "percentage"
	DiscountFixed      DiscountType = "fixed"
)

// IsValidDiscountType checks if a discount type is valid
func IsValidDiscountType(dt string) bool {
	switch DiscountType(dt) {
	case DiscountNone, DiscountPercentage, DiscountFixed:
		return true
	}
	return false
}

// CalculateDiscount calculates discount amount based on type and value
func CalculateDiscount(subtotal float64, discountType DiscountType, discountValue float64) float64 {
	if discountType == DiscountNone || discountValue <= 0 {
		return 0
	}

	if discountType == DiscountPercentage {
		// Cap percentage at 100%
		if discountValue > 100 {
			discountValue = 100
		}
		return subtotal * (discountValue / 100)
	}

	// Fixed discount - cap at subtotal
	if discountValue > subtotal {
		return subtotal
	}
	return discountValue
}

// CalculateTotal calculates total after discount
func CalculateTotal(subtotal float64, discountType DiscountType, discountValue float64) float64 {
	discount := CalculateDiscount(subtotal, discountType, discountValue)
	return subtotal - discount
}

// ValidateDiscount validates discount configuration
func ValidateDiscount(subtotal float64, discountType DiscountType, discountValue float64) error {
	if !IsValidDiscountType(string(discountType)) {
		return ErrInvalidDiscountType
	}

	if discountValue < 0 {
		return ErrNegativeDiscount
	}

	if discountType == DiscountPercentage && discountValue > 100 {
		return ErrDiscountPercentageExceeds100
	}

	if discountType == DiscountFixed && discountValue > subtotal {
		return ErrDiscountExceedsSubtotal
	}

	return nil
}

// Discount error types
type DiscountError string

const (
	ErrInvalidDiscountType       DiscountError = "invalid discount type"
	ErrNegativeDiscount          DiscountError = "discount cannot be negative"
	ErrDiscountPercentageExceeds100 DiscountError = "percentage discount cannot exceed 100%"
	ErrDiscountExceedsSubtotal   DiscountError = "fixed discount cannot exceed subtotal"
)

func (e DiscountError) Error() string {
	return string(e)
}
