package handlers

import (
	"debtNote/models"
	"debtNote/repository"
	"encoding/json"
	"net/http"
	"strconv"
)

// AddDebtRequest represents the incoming request for adding a debt.
type AddDebtRequest struct {
	Fullname  string  `json:"fullname"`
	Phone     string  `json:"phone"`
	Address   string  `json:"address"`
	PhotoData string  `json:"photo_data"`
	Amount    float64 `json:"amount"`
	Comment   string  `json:"comment"`
}

// PaginatedResponse is a generic wrapper for paginated data.
type PaginatedResponse struct {
	Data  interface{} `json:"data"`
	Total int         `json:"total"`
}

func GetDebtsHandler(w http.ResponseWriter, r *http.Request) {
	searchQuery := r.URL.Query().Get("search")
	dateQuery := r.URL.Query().Get("date")
	status := r.URL.Query().Get("status")
	
	clientID, _ := strconv.ParseInt(r.URL.Query().Get("client_id"), 10, 64)

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 20 // Default limit
	}


	debts, total, err := repository.GetDebts(searchQuery, dateQuery, status, clientID, page, limit)
	if err != nil {
		http.Error(w, "Failed to fetch debts: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := PaginatedResponse{
		Data:  debts,
		Total: total,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
	}
}

func AddDebtHandler(w http.ResponseWriter, r *http.Request) {
	var req AddDebtRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 1. Find or create the client
	client := models.Client{
		Fullname:  req.Fullname,
		Phone:     req.Phone,
		Address:   req.Address,
		PhotoData: req.PhotoData,
	}

	// Check if client exists first to determine if photo is required
	// Note: This logic is slightly simplified. Ideally, we should check if the client exists
	// and if not, THEN enforce the photo requirement.
	// However, repository.FindOrCreateClient handles both.
	// Let's modify repository.FindOrCreateClient to return an error if it's a new client and photo is missing.
	// Or better, let's check here if we can. Since we don't know if the client exists without querying DB,
	// we rely on the frontend validation mostly. But for backend safety:
	
	// We can try to find the client first.
	// But to keep it simple and robust: Let's modify FindOrCreateClient in repository to enforce photo for NEW clients.
	
	clientID, err := repository.FindOrCreateClient(client)
	if err != nil {
		http.Error(w, "Failed to process client: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 2. Add the debt for the client
	debt := models.Debt{
		ClientID: clientID,
		Amount:   req.Amount,
		Comment:  req.Comment,
	}
	_, err = repository.AddDebt(debt)
	if err != nil {
		http.Error(w, "Failed to add debt: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Debt added successfully"})
}

func PayDebtHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		DebtID int64             `json:"debt_id"`
		Rating models.DebtRating `json:"rating"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate rating
	switch payload.Rating {
	case models.RatingGood, models.RatingBad, models.RatingUntrusted:
		// valid
	default:
		http.Error(w, "Invalid rating value", http.StatusBadRequest)
		return
	}

	err := repository.PayDebt(payload.DebtID, payload.Rating)
	if err != nil {
		http.Error(w, "Failed to pay debt: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Debt paid successfully"})
}
