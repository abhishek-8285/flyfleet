package service

import (
	"math"
	"strings"
)

var (
	ones = []string{"", "One", "Two", "Three", "Four", "Five", "Six", "Seven", "Eight", "Nine", "Ten", "Eleven", "Twelve", "Thirteen", "Fourteen", "Fifteen", "Sixteen", "Seventeen", "Eighteen", "Nineteen"}
	tens = []string{"", "", "Twenty", "Thirty", "Forty", "Fifty", "Sixty", "Seventy", "Eighty", "Ninety"}
)

func AmountToWords(amount float64) string {
	if amount == 0 {
		return "Zero Rupees Only"
	}

	intPart := int64(math.Floor(amount))
	paisaPart := int64(math.Round((amount - float64(intPart)) * 100))

	words := convertIndian(intPart)
	if words == "" {
		words = "Zero"
	}

	result := words + " Rupees"

	if paisaPart > 0 {
		result += " and " + convertIndian(paisaPart) + " Paisa"
	}

	return result + " Only"
}

func convertIndian(n int64) string {
	if n == 0 {
		return ""
	}

	if n < 0 {
		return "Minus " + convertIndian(-n)
	}

	res := ""

	// Crores
	if n >= 10000000 {
		res += convertIndian(n/10000000) + " Crore "
		n %= 10000000
	}

	// Lakhs
	if n >= 100000 {
		res += convertIndian(n/100000) + " Lakh "
		n %= 100000
	}

	// Thousands
	if n >= 1000 {
		res += convertIndian(n/1000) + " Thousand "
		n %= 1000
	}

	// Hundreds
	if n >= 100 {
		res += convertIndian(n/100) + " Hundred "
		n %= 100
	}

	// Tens and Ones
	if n > 0 {
		if res != "" && n < 100 {
			// Optional "and" could be added here for British style, but not standard in Indian English legal docs
		}
		
		if n < 20 {
			res += ones[n]
		} else {
			res += tens[n/10]
			if n%10 > 0 {
				res += "-" + ones[n%10]
			}
		}
	}

	return strings.TrimSpace(res)
}
