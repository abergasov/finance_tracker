package entities

type HomePage struct {
	User                   *AuthUser     `json:"user"`
	SupportedCurrencies    []string      `json:"supported_currencies"`
	UserExpensesCategories *UserExpenses `json:"user_expenses_categories"`
}
