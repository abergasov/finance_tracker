package utils_test

import (
	"finance_tracker/internal/utils"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDeriveHexColorIsStable(t *testing.T) {
	got := utils.DeriveHexColor("  Groceries  ")
	require.Equal(t, got, utils.DeriveHexColor("groceries"))
	require.Regexp(t, `^#[0-9a-f]{6}$`, got)
}

func TestNormalizeHexColor(t *testing.T) {
	got, ok := utils.NormalizeHexColor("  #0F0f0F  ")
	require.True(t, ok)
	require.Equal(t, "#0f0f0f", got)

	table := map[string]struct{}{
		"not-a-color": {},
		"":            {},
	}
	for input := range table {
		_, ok = utils.NormalizeHexColor(input)
		require.False(t, ok)
	}
}
