package routes_test

import (
	"finance_tracker/internal/entities"
	testhelpers "finance_tracker/internal/test_helpers"
	"finance_tracker/internal/test_helpers/seed"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

type respID struct {
	ID int64 `json:"id"`
}

func TestCategoryCRUDRoutes(t *testing.T) {
	container := testhelpers.GetClean(t)
	srv := testhelpers.NewTestServerWithUser(t, container)

	home := loadHomePage(t, srv)
	mandatoryID := home.UserExpensesCategories.MandatoryExpenses.ID
	require.NotZero(t, mandatoryID)
	var createResp respID

	t.Run("should create a category", func(t *testing.T) {
		t.Run("should not create with invalid name", func(t *testing.T) {
			var payload map[string]string
			srv.Post(t, "/api/v1/category", map[string]any{
				"parent_id": mandatoryID,
				"name":      "   \t  ",
			}).RequireBadRequest(t).RequireUnmarshal(t, &payload)
			require.Equal(t, "name is required", payload["error"])

			homeAfter := loadHomePage(t, srv)
			require.Empty(t, homeAfter.UserExpensesCategories.MandatoryExpenses.Children)
			require.Empty(t, homeAfter.UserExpensesCategories.OptionalExpenses.Children)
		})
		// when
		srv.Post(t, "/api/v1/category", map[string]any{
			"parent_id": mandatoryID,
			"name":      "groceries",
		}).RequireCreated(t).RequireUnmarshal(t, &createResp)
		require.NotZero(t, createResp.ID)

		// then
		created := loadHomePage(t, srv).UserExpensesCategories.MandatoryExpenses.FindCategoryByID(createResp.ID)
		require.NotNil(t, created)
		require.Equal(t, "groceries", created.Name)
		t.Run("should reject on name duplicate", func(t *testing.T) {
			var payload map[string]string
			srv.Post(t, "/api/v1/category", map[string]any{
				"parent_id": mandatoryID,
				"name":      "groceries",
			}).RequireServerError(t).RequireUnmarshal(t, &payload)
			require.NotEmpty(t, payload["error"])
		})
	})

	t.Run("shuld able update an existing category", func(t *testing.T) {
		t.Run("should not allow update with invalid name", func(t *testing.T) {
			var payload map[string]string
			srv.Put(t, fmt.Sprintf("/api/v1/category/%d", createResp.ID), map[string]any{
				"name": " \n ",
			}).RequireBadRequest(t).RequireUnmarshal(t, &payload)
			require.Equal(t, "name is required", payload["error"])

			homeAfter := loadHomePage(t, srv)
			created := homeAfter.UserExpensesCategories.MandatoryExpenses.FindCategoryByID(createResp.ID)
			require.NotNil(t, created)
			require.Equal(t, "groceries", created.Name)
		})
		// when
		srv.Put(t, fmt.Sprintf("/api/v1/category/%d", createResp.ID), map[string]any{
			"name": "groceries-renamed",
		}).RequireOk(t)

		// then
		created := loadHomePage(t, srv).UserExpensesCategories.MandatoryExpenses.FindCategoryByID(createResp.ID)
		require.NotNil(t, created)
		require.Equal(t, "groceries-renamed", created.Name)
	})

	t.Run("should able to delete category", func(t *testing.T) {
		var grandResp respID
		srv.Post(t, "/api/v1/category", map[string]any{
			"parent_id": createResp.ID,
			"name":      "receipts",
		}).RequireCreated(t).RequireUnmarshal(t, &grandResp)
		require.NotZero(t, grandResp.ID)

		srv.Delete(t, fmt.Sprintf("/api/v1/category/%d", createResp.ID), nil).RequireOk(t)

		home = loadHomePage(t, srv)
		require.Nil(t, home.UserExpensesCategories.MandatoryExpenses.FindCategoryByID(createResp.ID))
		require.Nil(t, home.UserExpensesCategories.MandatoryExpenses.FindCategoryByID(grandResp.ID))
		require.Empty(t, home.UserExpensesCategories.MandatoryExpenses.Children)
	})
}

func TestCategoryRouteRejectsCrossUserParent(t *testing.T) {
	container := testhelpers.GetClean(t)
	srv := testhelpers.NewTestServer(t, container)

	owner := seed.NewUserBuilder().PopulateTest(t, container.Repo)
	other := seed.NewUserBuilder().PopulateTest(t, container.Repo)

	srv.AuthUser(other.Email)
	otherHome := loadHomePage(t, srv)
	otherRootID := otherHome.UserExpensesCategories.MandatoryExpenses.ID
	require.NotZero(t, otherRootID)

	srv.AuthUser(owner.Email)
	var payload map[string]string
	srv.Post(t, "/api/v1/category", map[string]any{
		"parent_id": otherRootID,
		"name":      "intruder",
	}).RequireServerError(t).RequireUnmarshal(t, &payload)
	require.NotEmpty(t, payload["error"])
}

func TestCategoryRouteTrimsPersistedNames(t *testing.T) {
	container := testhelpers.GetClean(t)
	srv := testhelpers.NewTestServerWithUser(t, container)

	home := loadHomePage(t, srv)
	parentID := home.UserExpensesCategories.MandatoryExpenses.ID

	var createResp respID
	srv.Post(t, "/api/v1/category", map[string]any{
		"parent_id": parentID,
		"name":      "  foo  ",
	}).RequireCreated(t).RequireUnmarshal(t, &createResp)

	homeAfter := loadHomePage(t, srv)
	created := homeAfter.UserExpensesCategories.MandatoryExpenses.FindCategoryByID(createResp.ID)
	require.NotNil(t, created)
	require.Equal(t, "foo", created.Name)
}

func TestCategoryRouteRejectsRenameConflict(t *testing.T) {
	container := testhelpers.GetClean(t)
	srv := testhelpers.NewTestServerWithUser(t, container)

	home := loadHomePage(t, srv)
	parentID := home.UserExpensesCategories.MandatoryExpenses.ID

	var firstResp respID
	srv.Post(t, "/api/v1/category", map[string]any{
		"parent_id": parentID,
		"name":      "alpha",
	}).RequireCreated(t).RequireUnmarshal(t, &firstResp)

	var secondResp respID
	srv.Post(t, "/api/v1/category", map[string]any{
		"parent_id": parentID,
		"name":      "beta",
	}).RequireCreated(t).RequireUnmarshal(t, &secondResp)

	var payload map[string]string
	srv.Put(t, fmt.Sprintf("/api/v1/category/%d", secondResp.ID), map[string]any{
		"name": "alpha",
	}).RequireServerError(t).RequireUnmarshal(t, &payload)
	require.NotEmpty(t, payload["error"])

	homeAfter := loadHomePage(t, srv)
	first := homeAfter.UserExpensesCategories.MandatoryExpenses.FindCategoryByID(firstResp.ID)
	second := homeAfter.UserExpensesCategories.MandatoryExpenses.FindCategoryByID(secondResp.ID)
	require.NotNil(t, first)
	require.NotNil(t, second)
	require.Equal(t, "alpha", first.Name)
	require.Equal(t, "beta", second.Name)
}

func TestCategoryRouteReturnsNotFoundForMissingCategory(t *testing.T) {
	container := testhelpers.GetClean(t)
	srv := testhelpers.NewTestServerWithUser(t, container)

	const missingID int64 = 999999999

	t.Run("put", func(t *testing.T) {
		srv.Put(t, fmt.Sprintf("/api/v1/category/%d", missingID), map[string]any{
			"name": "new-name",
		}).RequireOk(t)
	})

	t.Run("delete", func(t *testing.T) {
		srv.Delete(t, fmt.Sprintf("/api/v1/category/%d", missingID), nil).RequireOk(t)
	})
}

func TestCategoryCORSAllowsMutationHeaders(t *testing.T) {
	container := testhelpers.GetClean(t)
	srv := testhelpers.NewTestServer(t, container)

	response := srv.Request(t, http.MethodOptions, "/api/v1/category", nil, map[string]string{
		"Origin":                         container.Cfg.Auth.UIBaseURL,
		"Access-Control-Request-Method":  http.MethodPost,
		"Access-Control-Request-Headers": "authorization, content-type",
	}, nil).RequireNoContent(t)

	require.Equal(t, container.Cfg.Auth.UIBaseURL, response.Res.Header.Get("Access-Control-Allow-Origin"))
	require.Contains(t, response.Res.Header.Get("Access-Control-Allow-Methods"), http.MethodPost)
	require.Contains(t, response.Res.Header.Get("Access-Control-Allow-Headers"), "Authorization")
	require.Contains(t, response.Res.Header.Get("Access-Control-Allow-Headers"), "Content-Type")
	require.Contains(t, response.Res.Header.Get("Vary"), "Origin")
}

func TestCategoryRouteRejectsRootMutations(t *testing.T) {
	container := testhelpers.GetClean(t)
	srv := testhelpers.NewTestServerWithUser(t, container)

	home := loadHomePage(t, srv)
	mandatoryID := home.UserExpensesCategories.MandatoryExpenses.ID
	optionalID := home.UserExpensesCategories.OptionalExpenses.ID

	t.Run("reject root update", func(t *testing.T) {
		// when
		srv.Put(t, fmt.Sprintf("/api/v1/category/%d", mandatoryID), map[string]any{
			"name": "not-allowed",
		}).RequireOk(t)

		// then
		catID := loadHomePage(t, srv).UserExpensesCategories.MandatoryExpenses.FindCategoryByID(mandatoryID)
		require.Equal(t, catID.ID, mandatoryID)
	})

	t.Run("reject root delete", func(t *testing.T) {
		// when
		srv.Delete(t, fmt.Sprintf("/api/v1/category/%d", optionalID), nil).RequireOk(t)

		// then
		catID := loadHomePage(t, srv).UserExpensesCategories.OptionalExpenses.FindCategoryByID(optionalID)
		require.Equal(t, catID.ID, optionalID)
	})
}

func loadHomePage(t *testing.T, srv *testhelpers.TestServer) entities.HomePage {
	t.Helper()

	var home entities.HomePage
	srv.Get(t, "/api/v1/me").RequireOk(t).RequireUnmarshal(t, &home)
	return home
}
