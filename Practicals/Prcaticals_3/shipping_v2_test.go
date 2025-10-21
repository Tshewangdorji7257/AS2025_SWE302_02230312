// shipping_v2_test.go
package shipping

import (
	"strings"
	"testing"
)

func TestCalculateShippingFee_V2(t *testing.T) {
	testCases := []struct {
		name          string
		weight        float64
		zone          string
		insured       bool
		expectedFee   float64
		expectError   bool
		errorContains string
	}{
		// ===== EQUIVALENCE PARTITIONING TESTS =====
		
		// P1: Invalid Weight (Too Small)
		{
			name:          "P1: Invalid weight - negative",
			weight:        -5,
			zone:          "Domestic",
			insured:       false,
			expectedFee:   0,
			expectError:   true,
			errorContains: "invalid weight",
		},
		{
			name:          "P1: Invalid weight - zero",
			weight:        0,
			zone:          "International",
			insured:       true,
			expectedFee:   0,
			expectError:   true,
			errorContains: "invalid weight",
		},

		// P2: Standard Package (0 < weight <= 10)
		{
			name:          "P2: Standard package - mid range",
			weight:        5,
			zone:          "Domestic",
			insured:       false,
			expectedFee:   5.0, // Base: $5.00, No surcharge, No insurance
			expectError:   false,
		},
		{
			name:          "P2: Standard package - with insurance",
			weight:        8,
			zone:          "International",
			insured:       true,
			expectedFee:   20.3, // Base: $20.00, Insurance: $20.00 * 0.015 = $0.30
			expectError:   false,
		},

		// P3: Heavy Package (10 < weight <= 50)
		{
			name:          "P3: Heavy package - mid range",
			weight:        25,
			zone:          "Express",
			insured:       false,
			expectedFee:   37.5, // Base: $30.00, Surcharge: $7.50
			expectError:   false,
		},
		{
			name:          "P3: Heavy package - with insurance",
			weight:        15,
			zone:          "International",
			insured:       true,
			expectedFee:   27.9125, // Base: $20.00, Surcharge: $7.50, Insurance: $27.50 * 0.015 = $0.4125
			expectError:   false,
		},

		// P4: Invalid Weight (Too Large)
		{
			name:          "P4: Invalid weight - too large",
			weight:        100,
			zone:          "Domestic",
			insured:       false,
			expectedFee:   0,
			expectError:   true,
			errorContains: "invalid weight",
		},

		// P5: Valid Zones
		{
			name:          "P5: Valid zone - Domestic",
			weight:        10,
			zone:          "Domestic",
			insured:       false,
			expectedFee:   5.0,
			expectError:   false,
		},
		{
			name:          "P5: Valid zone - International",
			weight:        10,
			zone:          "International",
			insured:       false,
			expectedFee:   20.0,
			expectError:   false,
		},
		{
			name:          "P5: Valid zone - Express",
			weight:        10,
			zone:          "Express",
			insured:       false,
			expectedFee:   30.0,
			expectError:   false,
		},

		// P6: Invalid Zones
		{
			name:          "P6: Invalid zone - empty string",
			weight:        10,
			zone:          "",
			insured:       false,
			expectedFee:   0,
			expectError:   true,
			errorContains: "invalid zone",
		},
		{
			name:          "P6: Invalid zone - wrong case",
			weight:        10,
			zone:          "domestic",
			insured:       false,
			expectedFee:   0,
			expectError:   true,
			errorContains: "invalid zone",
		},
		{
			name:          "P6: Invalid zone - unknown",
			weight:        10,
			zone:          "Local",
			insured:       false,
			expectedFee:   0,
			expectError:   true,
			errorContains: "invalid zone",
		},

		// ===== BOUNDARY VALUE ANALYSIS TESTS =====

		// Lower Boundary (around 0)
		{
			name:          "BVA: Weight at lower invalid boundary",
			weight:        0,
			zone:          "Domestic",
			insured:       false,
			expectedFee:   0,
			expectError:   true,
			errorContains: "invalid weight",
		},
		{
			name:          "BVA: Weight just above lower boundary",
			weight:        0.1,
			zone:          "Domestic",
			insured:       false,
			expectedFee:   5.0,
			expectError:   false,
		},

		// Mid Boundary (around 10)
		{
			name:          "BVA: Weight at standard upper boundary",
			weight:        10,
			zone:          "International",
			insured:       false,
			expectedFee:   20.0, // No surcharge at exactly 10kg
			expectError:   false,
		},
		{
			name:          "BVA: Weight just above standard boundary",
			weight:        10.1,
			zone:          "International",
			insured:       false,
			expectedFee:   27.5, // Base: $20.00, Surcharge: $7.50
			expectError:   false,
		},

		// Upper Boundary (around 50)
		{
			name:          "BVA: Weight at upper valid boundary",
			weight:        50,
			zone:          "Express",
			insured:       false,
			expectedFee:   37.5, // Base: $30.00, Surcharge: $7.50
			expectError:   false,
		},
		{
			name:          "BVA: Weight just above upper boundary",
			weight:        50.1,
			zone:          "Express",
			insured:       false,
			expectedFee:   0,
			expectError:   true,
			errorContains: "invalid weight",
		},

		// ===== DECISION TABLE / COMBINATION TESTS =====

		// All valid combinations with insurance variations
		{
			name:          "Decision Table: Standard + Domestic + Insured",
			weight:        5,
			zone:          "Domestic",
			insured:       true,
			expectedFee:   5.075, // Base: $5.00, Insurance: $5.00 * 0.015 = $0.075
			expectError:   false,
		},
		{
			name:          "Decision Table: Heavy + International + Not Insured",
			weight:        20,
			zone:          "International",
			insured:       false,
			expectedFee:   27.5, // Base: $20.00, Surcharge: $7.50
			expectError:   false,
		},
		{
			name:          "Decision Table: Heavy + Express + Insured",
			weight:        30,
			zone:          "Express",
			insured:       true,
			expectedFee:   38.0625, // Base: $30.00, Surcharge: $7.50, Insurance: $37.50 * 0.015 = $0.5625
			expectError:   false,
		},

		// Edge case: Maximum values
		{
			name:          "Edge Case: Maximum weight + Express + Insured",
			weight:        50,
			zone:          "Express",
			insured:       true,
			expectedFee:   38.0625, // Base: $30.00, Surcharge: $7.50, Insurance: $37.50 * 0.015 = $0.5625
			expectError:   false,
		},

		// Edge case: Minimum valid values
		{
			name:          "Edge Case: Minimum weight + cheapest zone",
			weight:        0.1,
			zone:          "Domestic",
			insured:       false,
			expectedFee:   5.0,
			expectError:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fee, err := CalculateShippingFee(tc.weight, tc.zone, tc.insured)

			// Check error expectations
			if tc.expectError {
				if err == nil {
					t.Fatalf("Expected error containing '%s', but got nil", tc.errorContains)
				}
				if tc.errorContains != "" && !strings.Contains(err.Error(), tc.errorContains) {
					t.Errorf("Expected error to contain '%s', but got '%s'", tc.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, but got: %v", err)
				}
				// Check fee calculation (with small tolerance for floating point precision)
				tolerance := 0.0001
				if abs(fee-tc.expectedFee) > tolerance {
					t.Errorf("Expected fee %f, but got %f", tc.expectedFee, fee)
				}
			}
		})
	}
}

// Helper function to calculate absolute value
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}