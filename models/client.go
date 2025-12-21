package models

import "time"

type Client struct {
	ID        int64     `json:"id"`
	Fullname  string    `json:"fullname"`
	Phone     string    `json:"phone"`
	Address   string    `json:"address"`
	PhotoData string    `json:"photo_data"`
	CreatedAt time.Time `json:"created_at"`
}

// ClientSearchInfo represents a client in the search dropdown.
type ClientSearchInfo struct {
	ID            int64  `json:"id"`
	Fullname      string `json:"fullname"`
	Phone         string `json:"phone"`
	Address       string `json:"address"`
	PhotoData     string `json:"photo_data"`
	HasActiveDebt bool   `json:"has_active_debt"`
	Reputation    string `json:"reputation"` // 'untrusted', 'bad', 'good', or 'none'
}
