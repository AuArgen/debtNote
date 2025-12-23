package handlers

import (
	"debtNote/models"
	"debtNote/repository"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"
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
	sortBy := r.URL.Query().Get("sort_by")

	clientID, _ := strconv.ParseInt(r.URL.Query().Get("client_id"), 10, 64)

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 20 // Default limit
	}

	debts, total, err := repository.GetDebts(searchQuery, dateQuery, status, clientID, sortBy, page, limit)
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

	// Handle Photo Saving
	var photoPath string
	if req.PhotoData != "" {
		// Check if it's an existing file path (starts with /uploads/)
		if strings.HasPrefix(req.PhotoData, "/uploads/") {
			photoPath = req.PhotoData
		} else {
			// It's a new Base64 image, save it
			var err error
			photoPath, err = saveImage(req.PhotoData, req.Fullname)
			if err != nil {
				http.Error(w, "Failed to save image: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	// 1. Find or create the client
	client := models.Client{
		Fullname:  req.Fullname,
		Phone:     req.Phone,
		Address:   req.Address,
		PhotoData: photoPath, // Save the path
	}

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

// MakePaymentHandler handles partial or full payments.
func MakePaymentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		DebtID     int64             `json:"debt_id"`
		PaidAmount float64           `json:"paid_amount"`
		Comment    string            `json:"comment"`
		Rating     models.DebtRating `json:"rating"` // Used only if debt is fully paid
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if payload.Comment == "" {
		http.Error(w, "Комментарий милдеттүү", http.StatusBadRequest)
		return
	}

	err := repository.MakePayment(payload.DebtID, payload.PaidAmount, payload.Comment, payload.Rating)
	if err != nil {
		http.Error(w, "Failed to make payment: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Payment made successfully"})
}

// GetDebtPaymentsHandler retrieves payment history for a debt.
func GetDebtPaymentsHandler(w http.ResponseWriter, r *http.Request) {
	debtIDStr := r.URL.Query().Get("debt_id")
	debtID, err := strconv.ParseInt(debtIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid debt_id", http.StatusBadRequest)
		return
	}

	payments, err := repository.GetDebtPayments(debtID)
	if err != nil {
		http.Error(w, "Failed to get payments: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payments)
}

// DeleteDebtHandler handles soft deletion of a debt.
func DeleteDebtHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		DebtID  int64  `json:"debt_id"`
		Comment string `json:"comment"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if payload.Comment == "" {
		http.Error(w, "Өчүрүү себеби (комментарий) милдеттүү", http.StatusBadRequest)
		return
	}

	err := repository.DeleteDebt(payload.DebtID, payload.Comment)
	if err != nil {
		http.Error(w, "Failed to delete debt: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Debt deleted successfully"})
}

// saveImage decodes base64 image and saves it to disk
func saveImage(base64Data, fullname string) (string, error) {
	// Remove the data URL prefix (e.g., "data:image/jpeg;base64,")
	parts := strings.Split(base64Data, ",")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid base64 data")
	}

	data, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", err
	}

	// Create directory for current month: uploads/YYYY-MM
	currentMonth := time.Now().Format("2006-01")
	dirPath := filepath.Join("uploads", currentMonth)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", err
	}

	// Sanitize fullname for filename (Allow letters, numbers, spaces, underscores, hyphens)
	// We use unicode.IsLetter to support Cyrillic and other scripts
	safeName := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_' || r == '-' {
			return r
		}
		if unicode.IsSpace(r) {
			return '_'
		}
		return -1 // Drop other characters
	}, fullname)

	// Fallback if name becomes empty
	if safeName == "" {
		safeName = "unknown"
	}

	// Generate filename: Name_Date_Random.jpg
	dateStr := time.Now().Format("2006-01-02")
	randNum := rand.Intn(100000)
	filename := fmt.Sprintf("%s_%s_%d.jpg", safeName, dateStr, randNum)
	filePath := filepath.Join(dirPath, filename)

	// Write file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", err
	}

	// Return the path relative to the server root, using forward slashes for URL compatibility
	// e.g., /uploads/2023-10/John_Doe_2023-10-27_123.jpg
	return "/" + filepath.ToSlash(filePath), nil
}
