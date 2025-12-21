package handlers

import (
	"debtNote/repository"
	"encoding/json"
	"net/http"
	"strconv"
)

// PaginatedResponse is defined in debt_handler.go, but we can redefine or reuse.
// Since they are in the same package, we can reuse it if it's exported, or define a local struct.
// Let's reuse the one from debt_handler.go since they are in the same package 'handlers'.

func SearchClientsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	clients, err := repository.SearchClients(query)
	if err != nil {
		http.Error(w, "Failed to search clients: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(clients); err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
	}
}

func GetClientsHandler(w http.ResponseWriter, r *http.Request) {
	searchQuery := r.URL.Query().Get("search")
	dateQuery := r.URL.Query().Get("date")

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 200 // Default limit
	}

	clients, total, err := repository.GetClients(searchQuery, dateQuery, page, limit)
	if err != nil {
		http.Error(w, "Failed to fetch clients: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := PaginatedResponse{
		Data:  clients,
		Total: total,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
	}
}
