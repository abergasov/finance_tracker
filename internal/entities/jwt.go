package entities

type SignedTokenClaims struct {
	Kind   string `json:"kind"`
	Exp    int64  `json:"exp"`
	UserID string `json:"user_id,omitempty"`
	Email  string `json:"email,omitempty"`
	Locale string `json:"locale,omitempty"`
	Name   string `json:"name,omitempty"`
	Nonce  string `json:"nonce,omitempty"`
}
