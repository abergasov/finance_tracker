package utils_test

import (
	"finance_tracker/internal/utils"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseAmountToMinor(t *testing.T) {
	tests := map[string]struct {
		input   string
		want    int64
		wantErr string
	}{
		"integer only":        {input: "10", want: 1000},
		"two decimal places":  {input: "12.34", want: 1234},
		"one decimal place":   {input: "5.5", want: 550},
		"two zeros after dot": {input: "0.01", want: 1},
		"leading whitespace":  {input: "  7.50  ", want: 750},
		"large amount":        {input: "9999999.99", want: 999999999},
		// overflow: major > MaxInt64/100 (92233720368547758)
		"overflow major": {input: "92233720368547759", wantErr: "amount too large"},
		// overflow: major == MaxInt64/100 but minor > MaxInt64%100 (7)
		"overflow boundary": {input: "92233720368547758.08", wantErr: "amount too large"},
		// just inside boundary: 92233720368547758.07 == MaxInt64 in minor units
		"max boundary ok":      {input: "92233720368547758.07", want: math.MaxInt64},
		"empty string":         {input: "", wantErr: "amount is required"},
		"whitespace only":      {input: "   ", wantErr: "amount is required"},
		"negative":             {input: "-5.00", wantErr: "amount must be positive"},
		"zero":                 {input: "0", wantErr: "amount must be greater than zero"},
		"zero decimal":         {input: "0.00", wantErr: "amount must be greater than zero"},
		"non-numeric":          {input: "abc", wantErr: "invalid amount format"},
		"three decimal places": {input: "10.123", wantErr: "amount may have at most 2 decimal places"},
		"trailing dot only":    {input: "10.", wantErr: "amount may have at most 2 decimal places"},
		"leading dot":          {input: ".50", wantErr: "invalid amount format"},
		"multiple dots":        {input: "1.2.3", wantErr: "invalid amount format"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := utils.ParseAmountToMinor(tt.input)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				require.Zero(t, got)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
