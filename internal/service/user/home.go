package user

import (
	"context"
	"database/sql"
	"finance_tracker/internal/entities"
	"finance_tracker/internal/utils"
	"fmt"
	"sort"

	"github.com/google/uuid"
)

var (
	usrExpensesCategories = utils.NewRWMap[uuid.UUID, *entities.UserExpenses]()
)

func (s *Service) ServeHomePage(ctx context.Context, token string) (*entities.HomePage, error) {
	usr, err := utils.ParseAuthToken(token)
	if err != nil {
		return nil, err
	}

	uID, err := uuid.Parse(usr.ID)
	if err != nil {
		return nil, err
	}
	userCategories, ok := usrExpensesCategories.Load(uID)
	if !ok {
		userCategories, err = s.buildUserExpensesCategories(ctx, uID)
		if err != nil {
			return nil, err
		}
		usrExpensesCategories.Store(uID, userCategories)
	}

	return &entities.HomePage{
		User:                   usr,
		SupportedCurrencies:    entities.KnownCurrencies,
		UserExpensesCategories: userCategories,
	}, nil
}

func (s *Service) buildUserExpensesCategories(ctx context.Context, uID uuid.UUID) (*entities.UserExpenses, error) {
	if err := s.repo.EnsureRootCategories(ctx, uID); err != nil {
		return nil, err
	}
	expensesList, err := s.repo.LoadAllUserExpenses(ctx, uID)
	if err != nil {
		return nil, err
	}
	if len(expensesList) == 0 {
		return &entities.UserExpenses{}, nil
	}
	return BuildUserExpensesTree(expensesList)
}

func BuildUserExpensesTree(rows []*entities.UserExpensesCategoryDB) (*entities.UserExpenses, error) {
	nodes := make(map[int64]*entities.UserExpensesCategory, len(rows))
	parents := make(map[int64]sql.NullInt64, len(rows))

	for _, row := range rows {
		if row == nil {
			continue
		}

		if _, exists := nodes[row.ID]; exists {
			return nil, fmt.Errorf("duplicate expense category id: %d", row.ID)
		}

		nodes[row.ID] = &entities.UserExpensesCategory{
			ID:   row.ID,
			Name: row.Name,
		}
		parents[row.ID] = row.ParentID
	}

	var roots []*entities.UserExpensesCategory

	for id, parentID := range parents {
		node := nodes[id]

		if !parentID.Valid {
			roots = append(roots, node)
			continue
		}

		parent, ok := nodes[parentID.Int64]
		if !ok {
			return nil, fmt.Errorf("missing parent category: id=%d parent_id=%d", id, parentID.Int64)
		}

		parent.Children = append(parent.Children, node)
	}

	// Sort children by ID for deterministic output.
	for _, node := range nodes {
		if len(node.Children) > 1 {
			sort.Slice(node.Children, func(i, j int) bool {
				return node.Children[i].ID < node.Children[j].ID
			})
		}
	}

	var res entities.UserExpenses

	for _, root := range roots {
		switch root.Name {
		case "mandatory":
			if res.MandatoryExpenses.ID != 0 {
				return nil, fmt.Errorf("duplicate root category: mandatory")
			}
			res.MandatoryExpenses = *root
		case "optional":
			if res.OptionalExpenses.ID != 0 {
				return nil, fmt.Errorf("duplicate root category: optional")
			}
			res.OptionalExpenses = *root
		default:
			return nil, fmt.Errorf("unknown root expense category: %s", root.Name)
		}
	}

	return &res, nil
}
