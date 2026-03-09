package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/flyfleet/fleet-core-service/internal/domains/governance/models"
)

type MaskingService interface {
	MaskData(data *models.DSARData, rules json.RawMessage) *models.DSARData
}

type maskingService struct{}

func NewMaskingService() MaskingService {
	return &maskingService{}
}

func (s *maskingService) MaskData(data *models.DSARData, rules json.RawMessage) *models.DSARData {
	var ruleMap map[string]string
	if err := json.Unmarshal(rules, &ruleMap); err != nil {
		return data
	}

	// Recursively apply masking rules to all fields in the data structure
	maskedData := *data
	
	// Apply to User Profile
	if data.UserProfile != nil {
		maskedData.UserProfile = s.maskRecursive(data.UserProfile, ruleMap)
	}
	
	// Apply to other fields (they are slices of interface{})
	maskedData.Vehicles = s.maskSlice(data.Vehicles, ruleMap)
	maskedData.TripHistory = s.maskSlice(data.TripHistory, ruleMap)
	maskedData.Telemetry = s.maskSlice(data.Telemetry, ruleMap)
	maskedData.FinanceRecords = s.maskSlice(data.FinanceRecords, ruleMap)
	maskedData.FuelLogs = s.maskSlice(data.FuelLogs, ruleMap)
	maskedData.GeofenceEvents = s.maskSlice(data.GeofenceEvents, ruleMap)
	maskedData.AuditLogs = s.maskSlice(data.AuditLogs, ruleMap)

	return &maskedData
}

func (s *maskingService) maskSlice(slice interface{}, rules map[string]string) interface{} {
	if slice == nil {
		return nil
	}
	sArr, ok := slice.([]interface{})
	if !ok {
		return slice
	}
	
	result := make([]interface{}, len(sArr))
	for i, v := range sArr {
		result[i] = s.maskRecursive(v, rules)
	}
	return result
}

func (s *maskingService) maskRecursive(val interface{}, rules map[string]string) interface{} {
	switch v := val.(type) {
	case map[string]interface{}:
		newMap := make(map[string]interface{})
		for k, val := range v {
			if maskRule, ok := rules[k]; ok {
				newMap[k] = s.applyRule(fmt.Sprintf("%v", val), maskRule)
			} else {
				newMap[k] = s.maskRecursive(val, rules)
			}
		}
		return newMap
	case []interface{}:
		newSlice := make([]interface{}, len(v))
		for i, item := range v {
			newSlice[i] = s.maskRecursive(item, rules)
		}
		return newSlice
	default:
		return v
	}
}

func (s *maskingService) applyRule(val, rule string) string {
	if val == "" {
		return val
	}
	
	// Implementation of rules like {last4}, {first3} etc.
	if strings.Contains(rule, "{last4}") {
		if len(val) > 4 {
			return strings.Replace(rule, "{last4}", val[len(val)-4:], 1)
		}
		return rule
	}
	
	if strings.Contains(rule, "{first}") && strings.Contains(rule, "{domain}") {
		parts := strings.Split(val, "@")
		if len(parts) == 2 {
			r := strings.Replace(rule, "{first}", parts[0][:1], 1)
			return strings.Replace(r, "{domain}", parts[1], 1)
		}
	}
	
	// Default to just returning the rule if it doesn't match placeholders or simple masking
	return rule
}
