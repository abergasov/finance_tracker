package routes

import (
	"encoding/json"
	"errors"
	"finance_tracker/internal/utils"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"finance_tracker/internal/entities"
	currencyservice "finance_tracker/internal/service/currency"
	userservice "finance_tracker/internal/service/user"
)

type createExpenseRequest struct {
	CategoryID int64  `json:"category_id"`
	Currency   string `json:"currency"`
	Amount     string `json:"amount"`
}

// handleCreateExpense handles POST /api/v1/expense.
// Requires an authenticated bearer token.
// Validates the category belongs to the user, is non-root, the currency is
// supported, and the amount is a positive decimal with at most 2 decimal places.
func (s *Router) handleCreateExpense(ctx fiber.Ctx, userID uuid.UUID) error {
	var req createExpenseRequest
	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.CategoryID <= 0 {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "category_id is required"})
	}

	currency, err := entities.CurrencyFromString(req.Currency)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "unsupported currency"})
	}

	req.Amount = strings.TrimSpace(req.Amount)
	amountMinor, err := utils.ParseAmountToMinor(req.Amount)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err = s.service.CreateExpenseRecord(ctx.Context(), userID, req.CategoryID, currency, amountMinor); err != nil {
		if errors.Is(err, currencyservice.ErrAmountTooLarge) {
			return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"error": currencyservice.ErrAmountTooLarge.Error()})
		}
		if errors.Is(err, userservice.ErrInvalidCategory) || errors.Is(err, userservice.ErrRootCategory) {
			return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return ctx.Status(http.StatusCreated).JSON(fiber.Map{"ok": true})
}
