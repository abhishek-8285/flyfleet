package pdf

import (
	"math"
	"strings"
)

// AmountToWords converts a float64 amount to Indian Rupees in words
func AmountToWords(amount float64) string {
	units := []string{"", "One", "Two", "Three", "Four", "Five", "Six", "Seven", "Eight", "Nine"}
	teens := []string{"Ten", "Eleven", "Twelve", "Thirteen", "Fourteen", "Fifteen", "Sixteen", "Seventeen", "Eighteen", "Nineteen"}
	tens := []string{"", "", "Twenty", "Thirty", "Forty", "Fifty", "Sixty", "Seventy", "Eighty", "Ninety"}

	convert := func(n int) string {
		if n == 0 {
			return ""
		}
		if n < 10 {
			return units[n]
		}
		if n < 20 {
			return teens[n-10]
		}
		if n < 100 {
			return tens[n/10] + " " + units[n%10]
		}
		return units[n/100] + " Hundred " + tens[(n%100)/10] + " " + units[n%10]
	}

	intPart := int(amount)
	paisePart := int(math.Round((amount - float64(intPart)) * 100))

	var result []string

	// Crores
	if intPart >= 10000000 {
		result = append(result, convert(intPart/10000000)+" Crore")
		intPart %= 10000000
	}
	// Lakhs
	if intPart >= 100000 {
		result = append(result, convert(intPart/100000)+" Lakh")
		intPart %= 100000
	}
	// Thousands
	if intPart >= 1000 {
		result = append(result, convert(intPart/1000)+" Thousand")
		intPart %= 1000
	}
	// Hundreds & below
	if intPart > 0 {
		result = append(result, convert(intPart))
	}

	final := strings.Join(result, " ") + " Rupees"
	if paisePart > 0 {
		final += " and " + convert(paisePart) + " Paise"
	}
	return final + " Only"
}
