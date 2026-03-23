package pdf

import (
	"fmt"
	"strings"
)

// DynamicFieldContract defines the "Expected" keys for a template
type DynamicFieldContract struct {
	RequiredKeys []string
	AllowedKeys  []string
}

// GlobalContract is the "Logistics Standard" for FlyFleet
var GlobalContract = DynamicFieldContract{
	RequiredKeys: []string{"transporter", "vehicle_no", "lr_no"},
	AllowedKeys:  []string{"transporter", "vehicle_no", "lr_no", "eway_bill", "place_of_supply", "reverse_charge", "driver_phone"},
}

// ValidateCustomFields ensures the operator didn't make a typo in the key names
func (c *DynamicFieldContract) ValidateCustomFields(fields map[string]interface{}) error {
	// 1. Check for Required Keys
	for _, req := range c.RequiredKeys {
		if _, ok := fields[req]; !ok {
			return fmt.Errorf("missing mandatory logistics field: %s", req)
		}
	}

	// 2. Check for Illegal/Typo Keys
	allowedMap := make(map[string]bool)
	for _, k := range c.AllowedKeys {
		allowedMap[k] = true
	}

	for k := range fields {
		if !allowedMap[strings.ToLower(k)] {
			return fmt.Errorf("invalid field name detected: '%s'. Did you mean one of %v?", k, c.AllowedKeys)
		}
	}

	return nil
}
