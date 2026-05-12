package routes_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"finance_tracker/internal/entities"
	currencyservice "finance_tracker/internal/service/currency"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	testhelpers "finance_tracker/internal/test_helpers"
	"finance_tracker/internal/test_helpers/seed"
)

func TestCreateExpenseRoute(t *testing.T) {
	// given
	container := testhelpers.GetClean(t)
	seedExpenseCurrencyRates(t, container)
	srv := testhelpers.NewTestServerWithUser(t, container)

	home := loadHomePage(t, srv)
	mandatoryID := home.UserExpensesCategories.MandatoryExpenses.ID
	require.NotZero(t, mandatoryID)

	// Create a child category to use as a valid non-root category.
	var catResp respID
	srv.Post(t, "/api/v1/category", map[string]any{
		"parent_id": mandatoryID,
		"name":      "groceries",
	}).RequireCreated(t).RequireUnmarshal(t, &catResp)
	childCategoryID := catResp.ID
	require.NotZero(t, childCategoryID)

	t.Run("creates expense with integer amount", func(t *testing.T) {
		// when
		var resp map[string]any
		srv.Post(t, "/api/v1/expense", map[string]any{
			"category_id": childCategoryID,
			"currency":    "USD",
			"amount":      "10",
		}).RequireCreated(t).RequireUnmarshal(t, &resp)

		// then
		require.Equal(t, true, resp["ok"])
	})

	t.Run("creates expense with decimal amount", func(t *testing.T) {
		// when
		var resp map[string]any
		srv.Post(t, "/api/v1/expense", map[string]any{
			"category_id": childCategoryID,
			"currency":    "EUR",
			"amount":      "12.34",
		}).RequireCreated(t).RequireUnmarshal(t, &resp)

		// then
		require.Equal(t, true, resp["ok"])
	})

	t.Run("creates expense with one decimal place", func(t *testing.T) {
		// when
		var resp map[string]any
		srv.Post(t, "/api/v1/expense", map[string]any{
			"category_id": childCategoryID,
			"currency":    "USD",
			"amount":      "5.5",
		}).RequireCreated(t).RequireUnmarshal(t, &resp)

		// then
		require.Equal(t, true, resp["ok"])
	})

	table := map[string]any{
		"rejects missing category_id": map[string]any{
			"currency": "USD",
			"amount":   "10.00",
		},
		"rejects zero category_id": map[string]any{
			"category_id": 0,
			"currency":    "USD",
			"amount":      "10.00",
		},
		"rejects unknown currency": map[string]any{
			"category_id": childCategoryID,
			"currency":    "NOPE",
			"amount":      "10.00",
		},
		"rejects empty amount": map[string]any{
			"category_id": childCategoryID,
			"currency":    "USD",
			"amount":      "",
		},
		"rejects zero amount": map[string]any{
			"category_id": childCategoryID,
			"currency":    "USD",
			"amount":      "0",
		},
		"rejects negative amount": map[string]any{
			"category_id": childCategoryID,
			"currency":    "USD",
			"amount":      "-5.00",
		},
		"rejects amount with more than 2 decimal places": map[string]any{
			"category_id": childCategoryID,
			"currency":    "USD",
			"amount":      "10.123",
		},
		"rejects non-numeric amount": map[string]any{
			"category_id": childCategoryID,
			"currency":    "USD",
			"amount":      "abc",
		},
		"rejects amount too large for USD conversion": map[string]any{
			"category_id": childCategoryID,
			"currency":    "BTC",
			"amount":      "92233720368547758.07",
		},
		"rejects root category": map[string]any{
			"category_id": mandatoryID,
			"currency":    "USD",
			"amount":      "10.00",
		},
		"rejects non-existent category": map[string]any{
			"category_id": 999999999,
			"currency":    "USD",
			"amount":      "10.00",
		},
	}
	for name, payload := range table {
		t.Run(name, func(t *testing.T) {
			var resp map[string]string
			srv.Post(t, "/api/v1/expense", payload).RequireBadRequest(t).RequireUnmarshal(t, &resp)
			require.NotEmpty(t, resp["error"])
		})
	}
}

func TestCreateExpenseRouteRejectsCrossUserCategory(t *testing.T) {
	container := testhelpers.GetClean(t)
	seedExpenseCurrencyRates(t, container)
	srv := testhelpers.NewTestServer(t, container)

	owner := seed.NewUserBuilder().PopulateTest(t, container.Repo)
	other := seed.NewUserBuilder().PopulateTest(t, container.Repo)

	// Create a category under owner's mandatory root.
	srv.AuthUser(owner.Email)
	ownerHome := loadHomePage(t, srv)
	var ownerCatResp respID
	srv.Post(t, "/api/v1/category", map[string]any{
		"parent_id": ownerHome.UserExpensesCategories.MandatoryExpenses.ID,
		"name":      "rent",
	}).RequireCreated(t).RequireUnmarshal(t, &ownerCatResp)

	// other user attempts to record an expense against owner's category.
	srv.AuthUser(other.Email)
	var resp map[string]string
	srv.Post(t, "/api/v1/expense", map[string]any{
		"category_id": ownerCatResp.ID,
		"currency":    "USD",
		"amount":      "50.00",
	}).RequireBadRequest(t).RequireUnmarshal(t, &resp)
	require.NotEmpty(t, resp["error"])
}

func TestCreateExpenseRouteCORSAllowsMutationHeaders(t *testing.T) {
	container := testhelpers.GetClean(t)
	seedExpenseCurrencyRates(t, container)
	srv := testhelpers.NewTestServer(t, container)

	response := srv.Request(t, http.MethodOptions, "/api/v1/expense", nil, map[string]string{
		"Origin":                         container.Cfg.Auth.UIBaseURL,
		"Access-Control-Request-Method":  http.MethodPost,
		"Access-Control-Request-Headers": "authorization, content-type",
	}, nil).RequireNoContent(t)

	require.Equal(t, container.Cfg.Auth.UIBaseURL, response.Res.Header.Get("Access-Control-Allow-Origin"))
	require.Contains(t, response.Res.Header.Get("Access-Control-Allow-Methods"), http.MethodPost)
}

func TestCreateExpenseRoutePersistedRow(t *testing.T) {
	container := testhelpers.GetClean(t)
	seedExpenseCurrencyRates(t, container)
	srv := testhelpers.NewTestServerWithUser(t, container)

	home := loadHomePage(t, srv)
	mandatoryID := home.UserExpensesCategories.MandatoryExpenses.ID
	require.NotZero(t, mandatoryID)

	var catResp respID
	srv.Post(t, "/api/v1/category", map[string]any{
		"parent_id": mandatoryID,
		"name":      "persistence-check",
	}).RequireCreated(t).RequireUnmarshal(t, &catResp)
	require.NotZero(t, catResp.ID)

	srv.Post(t, "/api/v1/expense", map[string]any{
		"category_id": catResp.ID,
		"currency":    "EUR",
		"amount":      "12.34",
	}).RequireCreated(t)

	uID, err := uuid.Parse(home.User.ID)
	require.NoError(t, err)
	rows, err := container.Repo.LoadExpenseRecords(container.Ctx, uID)
	require.NoError(t, err)
	require.Len(t, rows, 1)

	require.Equal(t, catResp.ID, rows[0].CategoryID)
	require.Equal(t, "EUR", rows[0].Currency)
	require.Equal(t, int64(1234), rows[0].AmountMinor)
	require.Equal(t, int64(1142), rows[0].AmountMinorUSD)
}

func TestCreateExpenseWithGrandchildCategory(t *testing.T) {
	container := testhelpers.GetClean(t)
	seedExpenseCurrencyRates(t, container)
	srv := testhelpers.NewTestServerWithUser(t, container)

	home := loadHomePage(t, srv)
	mandatoryID := home.UserExpensesCategories.MandatoryExpenses.ID

	// child of mandatory
	var childResp respID
	srv.Post(t, "/api/v1/category", map[string]any{
		"parent_id": mandatoryID,
		"name":      "travel",
	}).RequireCreated(t).RequireUnmarshal(t, &childResp)

	// grandchild of mandatory (child of child)
	var grandResp respID
	srv.Post(t, "/api/v1/category", map[string]any{
		"parent_id": childResp.ID,
		"name":      "flights",
	}).RequireCreated(t).RequireUnmarshal(t, &grandResp)

	// Both child and grandchild should be accepted as non-root categories.
	for _, desc := range []struct {
		label string
		catID int64
	}{
		{"child", childResp.ID},
		{"grandchild", grandResp.ID},
	} {
		t.Run(fmt.Sprintf("accepts %s category", desc.label), func(t *testing.T) {
			var resp map[string]any
			srv.Post(t, "/api/v1/expense", map[string]any{
				"category_id": desc.catID,
				"currency":    "USD",
				"amount":      "100.00",
			}).RequireCreated(t).RequireUnmarshal(t, &resp)
			require.Equal(t, true, resp["ok"])
		})
	}
}

func TestCreateExpenseRouteRejectsWhenRatesUnavailable(t *testing.T) {
	container := testhelpers.GetClean(t)
	srv := testhelpers.NewTestServerWithUser(t, container)

	home := loadHomePage(t, srv)
	var catResp respID
	srv.Post(t, "/api/v1/category", map[string]any{
		"parent_id": home.UserExpensesCategories.MandatoryExpenses.ID,
		"name":      "missing-rates-check",
	}).RequireCreated(t).RequireUnmarshal(t, &catResp)

	var resp map[string]string
	srv.Post(t, "/api/v1/expense", map[string]any{
		"category_id": catResp.ID,
		"currency":    "EUR",
		"amount":      "12.34",
	}).RequireStatus(t, http.StatusInternalServerError).RequireUnmarshal(t, &resp)
	require.NotEmpty(t, resp["error"])

	uID, err := uuid.Parse(home.User.ID)
	require.NoError(t, err)
	rows, err := container.Repo.LoadExpenseRecords(container.Ctx, uID)
	require.NoError(t, err)
	require.Empty(t, rows)
}

func TestCreateExpenseRouteAllowsUSDWithoutRates(t *testing.T) {
	container := testhelpers.GetClean(t)
	srv := testhelpers.NewTestServerWithUser(t, container)

	home := loadHomePage(t, srv)
	var catResp respID
	srv.Post(t, "/api/v1/category", map[string]any{
		"parent_id": home.UserExpensesCategories.MandatoryExpenses.ID,
		"name":      "usd-without-rates",
	}).RequireCreated(t).RequireUnmarshal(t, &catResp)

	srv.Post(t, "/api/v1/expense", map[string]any{
		"category_id": catResp.ID,
		"currency":    "USD",
		"amount":      "12.34",
	}).RequireCreated(t)

	uID, err := uuid.Parse(home.User.ID)
	require.NoError(t, err)
	rows, err := container.Repo.LoadExpenseRecords(container.Ctx, uID)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, int64(1234), rows[0].AmountMinorUSD)
}

func TestCreateExpenseRouteRejectsOverflowWithoutPersistingRow(t *testing.T) {
	container := testhelpers.GetClean(t)
	seedExpenseCurrencyRates(t, container)
	srv := testhelpers.NewTestServerWithUser(t, container)

	home := loadHomePage(t, srv)
	var catResp respID
	srv.Post(t, "/api/v1/category", map[string]any{
		"parent_id": home.UserExpensesCategories.MandatoryExpenses.ID,
		"name":      "overflow-check",
	}).RequireCreated(t).RequireUnmarshal(t, &catResp)

	var resp map[string]string
	srv.Post(t, "/api/v1/expense", map[string]any{
		"category_id": catResp.ID,
		"currency":    "BTC",
		"amount":      "92233720368547758.07",
	}).RequireBadRequest(t).RequireUnmarshal(t, &resp)
	require.Equal(t, "amount too large", resp["error"])

	uID, err := uuid.Parse(home.User.ID)
	require.NoError(t, err)
	rows, err := container.Repo.LoadExpenseRecords(container.Ctx, uID)
	require.NoError(t, err)
	require.Empty(t, rows)
}

func seedExpenseCurrencyRates(t *testing.T, container *testhelpers.TestContainer) {
	t.Helper()
	require.NoError(t, container.Repo.SaveCurrencyRates(container.Ctx, map[entities.Currency]int64{
		entities.CurrencyUSD: currencyservice.RateScale,
		entities.CurrencyEUR: 108000000,
		entities.CurrencyGBP: 126000000,
		entities.CurrencyJPY: 14900000000,
		entities.CurrencyBTC: 1,
	}, time.Now()))
}
