package routes

import (
	"encoding/json"
	"errors"
	"finance_tracker/internal/repository"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
)

type createCategoryRequest struct {
	ParentID int64  `json:"parent_id"`
	Name     string `json:"name"`
}

type updateCategoryRequest struct {
	Name string `json:"name"`
}

// categoryError maps domain-level category errors to stable HTTP responses.
// User-caused domain errors get specific 4xx codes; unexpected storage errors
// get a generic 500 so internal details are not leaked.
func categoryError(ctx fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, repository.ErrParentNotFound), errors.Is(err, repository.ErrCategoryNotFound):
		return ctx.Status(http.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, repository.ErrProtectedRoot):
		return ctx.Status(http.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, repository.ErrDuplicateName):
		return ctx.Status(http.StatusConflict).JSON(fiber.Map{"error": err.Error()})
	default:
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "internal error"})
	}
}

// handleCreateCategory handles POST /api/auth/category.
// Requires an authenticated bearer token. The parent_id must point to an
// existing category owned by the user; name must be non-empty.
func (s *Router) handleCreateCategory(ctx fiber.Ctx) error {
	userID, err := s.extractUserID(ctx)
	if err != nil {
		return ctx.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	var req createCategoryRequest
	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "name is required"})
	}
	if req.ParentID <= 0 {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "parent_id is required"})
	}

	id, err := s.service.AddUserExpense(ctx.Context(), userID, req.ParentID, req.Name)
	if err != nil {
		return categoryError(ctx, err)
	}
	return ctx.Status(http.StatusCreated).JSON(fiber.Map{"id": id})
}

// handleUpdateCategory handles PUT /api/auth/category/:id.
// Root categories (mandatory/optional) cannot be renamed.
func (s *Router) handleUpdateCategory(ctx fiber.Ctx) error {
	userID, err := s.extractUserID(ctx)
	if err != nil {
		return ctx.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	id, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid category id"})
	}

	var req updateCategoryRequest
	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "name is required"})
	}

	if err := s.service.UpdateUserExpense(ctx.Context(), userID, id, req.Name); err != nil {
		return categoryError(ctx, err)
	}
	return ctx.Status(http.StatusOK).JSON(fiber.Map{"ok": true})
}

// handleDeleteCategory handles DELETE /api/auth/category/:id.
// Root categories cannot be deleted. Deletion cascades to all descendants.
func (s *Router) handleDeleteCategory(ctx fiber.Ctx) error {
	userID, err := s.extractUserID(ctx)
	if err != nil {
		return ctx.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	id, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid category id"})
	}

	if err := s.service.DeleteUserExpense(ctx.Context(), userID, id); err != nil {
		return categoryError(ctx, err)
	}
	return ctx.Status(http.StatusOK).JSON(fiber.Map{"ok": true})
}
