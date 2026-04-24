package routes

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type updateCategoryRequest struct {
	Name string `json:"name"`
}

// handleCreateCategory handles POST /api/auth/category.
// Requires an authenticated bearer token. The parent_id must point to an
// existing category owned by the user; name must be non-empty.
func (s *Router) handleCreateCategory(ctx fiber.Ctx, userID uuid.UUID) error {
	type createCategoryRequest struct {
		ParentID int64  `json:"parent_id"`
		Name     string `json:"name"`
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
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return ctx.Status(http.StatusCreated).JSON(fiber.Map{"id": id})
}

// handleUpdateCategory handles PUT /api/auth/category/:id.
// Root categories (mandatory/optional) cannot be renamed.
func (s *Router) handleUpdateCategory(ctx fiber.Ctx, userID uuid.UUID) error {
	id, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid category id"})
	}

	var req updateCategoryRequest
	if err = json.Unmarshal(ctx.Body(), &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "name is required"})
	}

	if err = s.service.UpdateUserExpense(ctx.Context(), userID, id, req.Name); err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return ctx.Status(http.StatusOK).JSON(fiber.Map{"ok": true})
}

// handleDeleteCategory handles DELETE /api/auth/category/:id.
// Root categories cannot be deleted. Deletion cascades to all descendants.
func (s *Router) handleDeleteCategory(ctx fiber.Ctx, userID uuid.UUID) error {
	id, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid category id"})
	}

	if err = s.service.DeleteUserExpense(ctx.Context(), userID, id); err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
	return ctx.Status(http.StatusOK).JSON(fiber.Map{"ok": true})
}
