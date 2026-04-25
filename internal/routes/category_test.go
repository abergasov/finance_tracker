package routes_test

import (
	"finance_tracker/internal/entities"
	testhelpers "finance_tracker/internal/test_helpers"
	"finance_tracker/internal/test_helpers/seed"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCategoryCRUDRoutes(t *testing.T) {
	container := testhelpers.GetClean(t)
	srv := testhelpers.NewTestServer(t, container)

	usr := seed.NewUserBuilder().PopulateTest(t, container)
	token := signToken(t, container.Cfg.Auth.Token.SigningKey, &entities.SignedTokenClaims{
		Kind:   "auth",
		Exp:    time.Now().Add(time.Hour).Unix(),
		UserID: usr.ID.String(),
		Email:  usr.Email,
		Name:   usr.Name,
	})
	auth := withBearer(token)

	home := loadHomePage(t, srv, auth)
	mandatoryID := home.UserExpensesCategories.MandatoryExpenses.ID
	require.NotZero(t, mandatoryID)

	createResp := struct {
		ID int64 `json:"id"`
	}{}
	srv.Request(t, http.MethodPost, "/api/auth/category", map[string]any{
		"parent_id": mandatoryID,
		"name":      "groceries",
	}, auth, nil).RequireCreated(t).RequireUnmarshal(t, &createResp)
	require.NotZero(t, createResp.ID)

	home = loadHomePage(t, srv, auth)
	created := findCategoryByID(&home.UserExpensesCategories.MandatoryExpenses, createResp.ID)
	require.NotNil(t, created)
	require.Equal(t, "groceries", created.Name)

	srv.Request(t, http.MethodPut, fmt.Sprintf("/api/auth/category/%d", createResp.ID), map[string]any{
		"name": "groceries-renamed",
	}, auth, nil).RequireOk(t)

	home = loadHomePage(t, srv, auth)
	created = findCategoryByID(&home.UserExpensesCategories.MandatoryExpenses, createResp.ID)
	require.NotNil(t, created)
	require.Equal(t, "groceries-renamed", created.Name)

	grandResp := struct {
		ID int64 `json:"id"`
	}{}
	srv.Request(t, http.MethodPost, "/api/auth/category", map[string]any{
		"parent_id": createResp.ID,
		"name":      "receipts",
	}, auth, nil).RequireCreated(t).RequireUnmarshal(t, &grandResp)
	require.NotZero(t, grandResp.ID)

	srv.Request(t, http.MethodDelete, fmt.Sprintf("/api/auth/category/%d", createResp.ID), nil, auth, nil).RequireOk(t)

	home = loadHomePage(t, srv, auth)
	require.Nil(t, findCategoryByID(&home.UserExpensesCategories.MandatoryExpenses, createResp.ID))
	require.Nil(t, findCategoryByID(&home.UserExpensesCategories.MandatoryExpenses, grandResp.ID))
	require.Empty(t, home.UserExpensesCategories.MandatoryExpenses.Children)
}

func TestCategoryRouteRejectsCrossUserParent(t *testing.T) {
	container := testhelpers.GetClean(t)
	srv := testhelpers.NewTestServer(t, container)

	owner := seed.NewUserBuilder().PopulateTest(t, container)
	other := seed.NewUserBuilder().PopulateTest(t, container)

	ownerToken := signToken(t, container.Cfg.Auth.Token.SigningKey, &entities.SignedTokenClaims{
		Kind:   "auth",
		Exp:    time.Now().Add(time.Hour).Unix(),
		UserID: owner.ID.String(),
		Email:  owner.Email,
		Name:   owner.Name,
	})
	otherToken := signToken(t, container.Cfg.Auth.Token.SigningKey, &entities.SignedTokenClaims{
		Kind:   "auth",
		Exp:    time.Now().Add(time.Hour).Unix(),
		UserID: other.ID.String(),
		Email:  other.Email,
		Name:   other.Name,
	})

	otherHome := loadHomePage(t, srv, withBearer(otherToken))
	otherRootID := otherHome.UserExpensesCategories.MandatoryExpenses.ID
	require.NotZero(t, otherRootID)

	var payload map[string]string
	srv.Request(t, http.MethodPost, "/api/auth/category", map[string]any{
		"parent_id": otherRootID,
		"name":      "intruder",
	}, withBearer(ownerToken), nil).RequireNotFound(t).RequireUnmarshal(t, &payload)
	require.Equal(t, "parent category not found", payload["error"])
}

func TestCategoryRouteRejectsDuplicateSiblingName(t *testing.T) {
	container := testhelpers.GetClean(t)
	srv := testhelpers.NewTestServer(t, container)

	usr := seed.NewUserBuilder().PopulateTest(t, container)
	token := signToken(t, container.Cfg.Auth.Token.SigningKey, &entities.SignedTokenClaims{
		Kind:   "auth",
		Exp:    time.Now().Add(time.Hour).Unix(),
		UserID: usr.ID.String(),
		Email:  usr.Email,
		Name:   usr.Name,
	})
	auth := withBearer(token)

	home := loadHomePage(t, srv, auth)
	parentID := home.UserExpensesCategories.MandatoryExpenses.ID

	createResp := struct {
		ID int64 `json:"id"`
	}{}
	srv.Request(t, http.MethodPost, "/api/auth/category", map[string]any{
		"parent_id": parentID,
		"name":      "groceries",
	}, auth, nil).RequireCreated(t).RequireUnmarshal(t, &createResp)
	require.NotZero(t, createResp.ID)

	var payload map[string]string
	srv.Request(t, http.MethodPost, "/api/auth/category", map[string]any{
		"parent_id": parentID,
		"name":      "groceries",
	}, auth, nil).RequireConflict(t).RequireUnmarshal(t, &payload)
	require.Equal(t, "a category with this name already exists under this parent", payload["error"])
}

func TestCategoryRouteRejectsWhitespaceOnlyNames(t *testing.T) {
	container := testhelpers.GetClean(t)
	srv := testhelpers.NewTestServer(t, container)

	usr := seed.NewUserBuilder().PopulateTest(t, container)
	token := signToken(t, container.Cfg.Auth.Token.SigningKey, &entities.SignedTokenClaims{
		Kind:   "auth",
		Exp:    time.Now().Add(time.Hour).Unix(),
		UserID: usr.ID.String(),
		Email:  usr.Email,
		Name:   usr.Name,
	})
	auth := withBearer(token)

	home := loadHomePage(t, srv, auth)
	parentID := home.UserExpensesCategories.MandatoryExpenses.ID

	t.Run("post", func(t *testing.T) {
		var payload map[string]string
		srv.Request(t, http.MethodPost, "/api/auth/category", map[string]any{
			"parent_id": parentID,
			"name":      "   \t  ",
		}, auth, nil).RequireBadRequest(t).RequireUnmarshal(t, &payload)
		require.Equal(t, "name is required", payload["error"])

		homeAfter := loadHomePage(t, srv, auth)
		require.Empty(t, homeAfter.UserExpensesCategories.MandatoryExpenses.Children)
		require.Empty(t, homeAfter.UserExpensesCategories.OptionalExpenses.Children)
	})

	createResp := struct {
		ID int64 `json:"id"`
	}{}
	srv.Request(t, http.MethodPost, "/api/auth/category", map[string]any{
		"parent_id": parentID,
		"name":      "groceries",
	}, auth, nil).RequireCreated(t).RequireUnmarshal(t, &createResp)

	t.Run("put", func(t *testing.T) {
		var payload map[string]string
		srv.Request(t, http.MethodPut, fmt.Sprintf("/api/auth/category/%d", createResp.ID), map[string]any{
			"name": " \n ",
		}, auth, nil).RequireBadRequest(t).RequireUnmarshal(t, &payload)
		require.Equal(t, "name is required", payload["error"])

		homeAfter := loadHomePage(t, srv, auth)
		created := findCategoryByID(&homeAfter.UserExpensesCategories.MandatoryExpenses, createResp.ID)
		require.NotNil(t, created)
		require.Equal(t, "groceries", created.Name)
	})
}

func TestCategoryRouteTrimsPersistedNames(t *testing.T) {
	container := testhelpers.GetClean(t)
	srv := testhelpers.NewTestServer(t, container)

	usr := seed.NewUserBuilder().PopulateTest(t, container)
	token := signToken(t, container.Cfg.Auth.Token.SigningKey, &entities.SignedTokenClaims{
		Kind:   "auth",
		Exp:    time.Now().Add(time.Hour).Unix(),
		UserID: usr.ID.String(),
		Email:  usr.Email,
		Name:   usr.Name,
	})
	auth := withBearer(token)

	home := loadHomePage(t, srv, auth)
	parentID := home.UserExpensesCategories.MandatoryExpenses.ID

	createResp := struct {
		ID int64 `json:"id"`
	}{}
	srv.Request(t, http.MethodPost, "/api/auth/category", map[string]any{
		"parent_id": parentID,
		"name":      "  foo  ",
	}, auth, nil).RequireCreated(t).RequireUnmarshal(t, &createResp)

	homeAfter := loadHomePage(t, srv, auth)
	created := findCategoryByID(&homeAfter.UserExpensesCategories.MandatoryExpenses, createResp.ID)
	require.NotNil(t, created)
	require.Equal(t, "foo", created.Name)
}

func TestCategoryRouteRejectsRenameConflict(t *testing.T) {
	container := testhelpers.GetClean(t)
	srv := testhelpers.NewTestServer(t, container)

	usr := seed.NewUserBuilder().PopulateTest(t, container)
	token := signToken(t, container.Cfg.Auth.Token.SigningKey, &entities.SignedTokenClaims{
		Kind:   "auth",
		Exp:    time.Now().Add(time.Hour).Unix(),
		UserID: usr.ID.String(),
		Email:  usr.Email,
		Name:   usr.Name,
	})
	auth := withBearer(token)

	home := loadHomePage(t, srv, auth)
	parentID := home.UserExpensesCategories.MandatoryExpenses.ID

	firstResp := struct {
		ID int64 `json:"id"`
	}{}
	srv.Request(t, http.MethodPost, "/api/auth/category", map[string]any{
		"parent_id": parentID,
		"name":      "alpha",
	}, auth, nil).RequireCreated(t).RequireUnmarshal(t, &firstResp)

	secondResp := struct {
		ID int64 `json:"id"`
	}{}
	srv.Request(t, http.MethodPost, "/api/auth/category", map[string]any{
		"parent_id": parentID,
		"name":      "beta",
	}, auth, nil).RequireCreated(t).RequireUnmarshal(t, &secondResp)

	var payload map[string]string
	srv.Request(t, http.MethodPut, fmt.Sprintf("/api/auth/category/%d", secondResp.ID), map[string]any{
		"name": "alpha",
	}, auth, nil).RequireConflict(t).RequireUnmarshal(t, &payload)
	require.Equal(t, "a category with this name already exists under this parent", payload["error"])

	homeAfter := loadHomePage(t, srv, auth)
	first := findCategoryByID(&homeAfter.UserExpensesCategories.MandatoryExpenses, firstResp.ID)
	second := findCategoryByID(&homeAfter.UserExpensesCategories.MandatoryExpenses, secondResp.ID)
	require.NotNil(t, first)
	require.NotNil(t, second)
	require.Equal(t, "alpha", first.Name)
	require.Equal(t, "beta", second.Name)
}

func TestCategoryRouteReturnsNotFoundForMissingCategory(t *testing.T) {
	container := testhelpers.GetClean(t)
	srv := testhelpers.NewTestServer(t, container)

	usr := seed.NewUserBuilder().PopulateTest(t, container)
	token := signToken(t, container.Cfg.Auth.Token.SigningKey, &entities.SignedTokenClaims{
		Kind:   "auth",
		Exp:    time.Now().Add(time.Hour).Unix(),
		UserID: usr.ID.String(),
		Email:  usr.Email,
		Name:   usr.Name,
	})
	auth := withBearer(token)

	const missingID int64 = 999999999

	t.Run("put", func(t *testing.T) {
		var payload map[string]string
		srv.Request(t, http.MethodPut, fmt.Sprintf("/api/auth/category/%d", missingID), map[string]any{
			"name": "new-name",
		}, auth, nil).RequireNotFound(t).RequireUnmarshal(t, &payload)
		require.Equal(t, "category not found", payload["error"])
	})

	t.Run("delete", func(t *testing.T) {
		var payload map[string]string
		srv.Request(t, http.MethodDelete, fmt.Sprintf("/api/auth/category/%d", missingID), nil, auth, nil).RequireNotFound(t).RequireUnmarshal(t, &payload)
		require.Equal(t, "category not found", payload["error"])
	})
}

func TestCategoryCORSAllowsMutationHeaders(t *testing.T) {
	container := testhelpers.GetClean(t)
	srv := testhelpers.NewTestServer(t, container)

	response := srv.Request(t, http.MethodOptions, "/api/auth/category", nil, map[string]string{
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
	srv := testhelpers.NewTestServer(t, container)

	usr := seed.NewUserBuilder().PopulateTest(t, container)
	token := signToken(t, container.Cfg.Auth.Token.SigningKey, &entities.SignedTokenClaims{
		Kind:   "auth",
		Exp:    time.Now().Add(time.Hour).Unix(),
		UserID: usr.ID.String(),
		Email:  usr.Email,
		Name:   usr.Name,
	})
	auth := withBearer(token)

	home := loadHomePage(t, srv, auth)
	mandatoryID := home.UserExpensesCategories.MandatoryExpenses.ID
	optionalID := home.UserExpensesCategories.OptionalExpenses.ID

	t.Run("reject root update", func(t *testing.T) {
		var payload map[string]string
		srv.Request(t, http.MethodPut, fmt.Sprintf("/api/auth/category/%d", mandatoryID), map[string]any{
			"name": "not-allowed",
		}, auth, nil).RequireForbidden(t).RequireUnmarshal(t, &payload)
		require.Equal(t, "root category cannot be modified", payload["error"])
	})

	t.Run("reject root delete", func(t *testing.T) {
		var payload map[string]string
		srv.Request(t, http.MethodDelete, fmt.Sprintf("/api/auth/category/%d", optionalID), nil, auth, nil).RequireForbidden(t).RequireUnmarshal(t, &payload)
		require.Equal(t, "root category cannot be modified", payload["error"])
	})
}

func loadHomePage(t *testing.T, srv *testhelpers.TestServer, auth map[string]string) entities.HomePage {
	t.Helper()

	var home entities.HomePage
	srv.GetWithHeader(t, "/api/auth/me", auth).RequireOk(t).RequireUnmarshal(t, &home)
	return home
}

func findCategoryByID(cat *entities.UserExpensesCategory, id int64) *entities.UserExpensesCategory {
	if cat == nil {
		return nil
	}
	if cat.ID == id {
		return cat
	}
	for _, child := range cat.Children {
		if found := findCategoryByID(child, id); found != nil {
			return found
		}
	}
	return nil
}
