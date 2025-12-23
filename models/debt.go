package models

import "time"

// DebtStatus defines the status of a debt.
type DebtStatus string

const (
	StatusActive  DebtStatus = "active"  // Карыз төлөнө элек
	StatusPaid    DebtStatus = "paid"    // Карыз төлөндү
	StatusDeleted DebtStatus = "deleted" // Карыз өчүрүлдү (ката же жокко чыгаруу)
)

// DebtRating defines the rating for a closed debt.
type DebtRating string

const (
	RatingGood      DebtRating = "good"      // Жакшы
	RatingBad       DebtRating = "bad"       // Начар
	RatingUntrusted DebtRating = "untrusted" // Ишенич жок
)

type Debt struct {
	ID            int64      `json:"id"`
	ClientID      int64      `json:"client_id"`
	Amount        float64    `json:"amount"`
	Comment       string     `json:"comment"`
	Status        DebtStatus `json:"status"`
	Rating        DebtRating `json:"rating,omitempty"` // Only for paid debts
	CreatedAt     time.Time  `json:"created_at"`
	PaidAt        *time.Time `json:"paid_at,omitempty"`        // Time when the debt was paid
	DeletedAt     *time.Time `json:"deleted_at,omitempty"`     // Time when the debt was deleted
	DeleteComment string     `json:"delete_comment,omitempty"` // Reason for deletion
}

// DebtPayment represents a partial or full payment record.
type DebtPayment struct {
	ID              int64     `json:"id"`
	DebtID          int64     `json:"debt_id"`
	PaidAmount      float64   `json:"paid_amount"`
	RemainingAmount float64   `json:"remaining_amount"`
	Comment         string    `json:"comment"`
	CreatedAt       time.Time `json:"created_at"`
}
