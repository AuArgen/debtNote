package models

import "time"

// DebtStatus defines the status of a debt.
type DebtStatus string

const (
	StatusActive DebtStatus = "active" // Карыз төлөнө элек
	StatusPaid   DebtStatus = "paid"   // Карыз төлөндү
)

// DebtRating defines the rating for a closed debt.
type DebtRating string

const (
	RatingGood      DebtRating = "good"      // Жакшы
	RatingBad       DebtRating = "bad"       // Начар
	RatingUntrusted DebtRating = "untrusted" // Ишенич жок
)

type Debt struct {
	ID        int64      `json:"id"`
	ClientID  int64      `json:"client_id"`
	Amount    float64    `json:"amount"`
	Comment   string     `json:"comment"`
	Status    DebtStatus `json:"status"`
	Rating    DebtRating `json:"rating,omitempty"` // Only for paid debts
	CreatedAt time.Time  `json:"created_at"`
	PaidAt    *time.Time `json:"paid_at,omitempty"` // Time when the debt was paid
}
