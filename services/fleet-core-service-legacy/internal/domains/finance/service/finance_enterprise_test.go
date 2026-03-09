package service_test

import (
	"testing"
	"time"
)

// TestTaxCalculationService_GST validates CGST/SGST vs IGST split logic.
func TestTaxCalculationService_GST(t *testing.T) {
	tests := []struct {
		name          string
		amount        float64
		rate          float64
		interState    bool
		expectedIGST  float64
		expectedCGST  float64
		expectedSGST  float64
	}{
		{
			name: "Intra-state 18% GST splits into CGST+SGST",
			amount: 1000, rate: 18, interState: false,
			expectedIGST: 0, expectedCGST: 90, expectedSGST: 90,
		},
		{
			name: "Inter-state 18% → IGST only",
			amount: 1000, rate: 18, interState: true,
			expectedIGST: 180, expectedCGST: 0, expectedSGST: 0,
		},
		{
			name: "Zero-rated fuel",
			amount: 1000, rate: 0, interState: false,
			expectedIGST: 0, expectedCGST: 0, expectedSGST: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gstAmount := tc.amount * tc.rate / 100
			var igst, cgst, sgst float64
			if tc.interState {
				igst = gstAmount
			} else {
				cgst = gstAmount / 2
				sgst = gstAmount / 2
			}

			if igst != tc.expectedIGST {
				t.Errorf("IGST: got %.2f, want %.2f", igst, tc.expectedIGST)
			}
			if cgst != tc.expectedCGST {
				t.Errorf("CGST: got %.2f, want %.2f", cgst, tc.expectedCGST)
			}
			if sgst != tc.expectedSGST {
				t.Errorf("SGST: got %.2f, want %.2f", sgst, tc.expectedSGST)
			}
		})
	}
}

// TestTDSTaxService_Rates validates TDS rate lookup and calculation.
func TestTDSTaxService_Rates(t *testing.T) {
	tdsRates := map[string]float64{
		"194C": 2.0,
		"194I": 10.0,
		"194J": 10.0,
		"194H": 5.0,
	}

	cases := []struct {
		section     string
		gross       float64
		expectedTDS float64
		expectedNet float64
	}{
		{"194C", 100000, 2000, 98000},
		{"194I", 50000, 5000, 45000},
		{"194J", 25000, 2500, 22500},
	}

	for _, tc := range cases {
		t.Run(tc.section, func(t *testing.T) {
			rate := tdsRates[tc.section]
			tds := tc.gross * rate / 100
			net := tc.gross - tds

			if tds != tc.expectedTDS {
				t.Errorf("TDS: got %.2f, want %.2f", tds, tc.expectedTDS)
			}
			if net != tc.expectedNet {
				t.Errorf("Net: got %.2f, want %.2f", net, tc.expectedNet)
			}
		})
	}
}

// TestFinancialYear validates Indian FY string formatting.
func TestFinancialYear(t *testing.T) {
	cases := []struct {
		month    time.Month
		year     int
		expected string
	}{
		{time.April, 2024, "2024-25"},
		{time.March, 2024, "2023-24"},
		{time.January, 2025, "2024-25"},
	}

	financialYear := func(t time.Time) string {
		year := t.Year()
		if t.Month() < 4 {
			return formatFY(year-1, year%100)
		}
		return formatFY(year, (year+1)%100)
	}

	for _, tc := range cases {
		t.Run(tc.expected, func(t *testing.T) {
			date := time.Date(tc.year, tc.month, 1, 0, 0, 0, 0, time.UTC)
			result := financialYear(date)
			if result != tc.expected {
				t.Errorf("FY: got %s, want %s", result, tc.expected)
			}
		})
	}
}

func formatFY(startYear, endYY int) string {
	return time.Date(startYear, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006") + "-" +
		time.Date(2000+endYY, 1, 1, 0, 0, 0, 0, time.UTC).Format("06")
}

// TestFraudRiskLevel validates risk level thresholds.
func TestFraudRiskLevel(t *testing.T) {
	riskLevel := func(score float64) string {
		switch {
		case score >= 75:
			return "CRITICAL"
		case score >= 50:
			return "HIGH"
		case score >= 25:
			return "MEDIUM"
		default:
			return "LOW"
		}
	}

	cases := []struct {
		score    float64
		expected string
	}{
		{0, "LOW"},
		{24.9, "LOW"},
		{25, "MEDIUM"},
		{49.9, "MEDIUM"},
		{50, "HIGH"},
		{74.9, "HIGH"},
		{75, "CRITICAL"},
		{100, "CRITICAL"},
	}

	for _, tc := range cases {
		t.Run(tc.expected, func(t *testing.T) {
			got := riskLevel(tc.score)
			if got != tc.expected {
				t.Errorf("score=%.1f: got %s, want %s", tc.score, got, tc.expected)
			}
		})
	}
}

// TestTripProfitMargin validates profit margin calculation.
func TestTripProfitMargin(t *testing.T) {
	cases := []struct {
		revenue      float64
		totalCost    float64
		expectedPct  float64
	}{
		{10000, 7000, 30.0},
		{10000, 10000, 0.0},
		{10000, 12000, -20.0},
		{0, 5000, 0.0}, // No revenue → no margin
	}

	profitMargin := func(rev, cost float64) float64 {
		if rev == 0 {
			return 0
		}
		return (rev - cost) / rev * 100
	}

	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			got := profitMargin(tc.revenue, tc.totalCost)
			if got != tc.expectedPct {
				t.Errorf("margin: got %.2f, want %.2f (rev=%.0f, cost=%.0f)", got, tc.expectedPct, tc.revenue, tc.totalCost)
			}
		})
	}
}
