package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"finance_tracker/internal/entities"
	"finance_tracker/internal/utils"
)

const (
	TableCurrencies = "currency_rates"
)

var (
	tableCurrenciesCols = []string{
		"timestamp",
		"timestamp_num",
		"rates",
	}
	tableCurrenciesColsStr = strings.Join(tableCurrenciesCols, ",")
)

func (r *Repo) LoadAllCurrencies(ctx context.Context) ([]*entities.CurrencyDB, error) {
	q := fmt.Sprintf("SELECT %s FROM %s", tableCurrenciesColsStr, TableCurrencies)
	return utils.QueryRowsToStruct[entities.CurrencyDB](ctx, r.db.Client(), q)
}

func (r *Repo) SaveCurrencyRates(ctx context.Context, rates map[entities.Currency]int64, rateFor time.Time) error {
	ratesJSON, err := json.Marshal(rates)
	if err != nil {
		return fmt.Errorf("could not marshal currency rates: %w", err)
	}
	q, p := utils.GenerateInsertSQL(TableCurrencies, map[string]any{
		"timestamp":     rateFor,
		"timestamp_num": utils.TimeToDayIntNum(rateFor),
		"rates":         ratesJSON,
	})
	q += " ON CONFLICT (timestamp_num) DO UPDATE SET timestamp = EXCLUDED.timestamp, rates = EXCLUDED.rates;"
	_, err = r.db.Client().ExecContext(ctx, q, p...)
	if err != nil {
		return fmt.Errorf("could not save currency rates: %w", err)
	}
	return nil
}
