package main

import (
	"debtNote/database"
	"debtNote/handlers"
	"log"
	"net/http"
)

func main() {
	// Initialize database
	database.InitDB()
	defer database.DB.Close()

	// Serve static files
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// API routes
	http.HandleFunc("/api/clients", handlers.GetClientsHandler)
	http.HandleFunc("/api/clients/search", handlers.SearchClientsHandler)
	http.HandleFunc("/api/debts", handlers.GetDebtsHandler)
	http.HandleFunc("/api/debts/add", handlers.AddDebtHandler)
	http.HandleFunc("/api/debts/pay", handlers.PayDebtHandler)


	// Handle SPA (Single Page Application) routing
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// If the request is for an API endpoint, let the respective handler manage it.
		// Otherwise, serve the main index.html file. This allows frontend routing to work.
		if r.URL.Path == "/" || r.URL.Path == "/clients" || r.URL.Path == "/history" {
			http.ServeFile(w, r, "static/index.html")
		} else if len(r.URL.Path) > 1 && (r.URL.Path[1:4] == "api") {
			// This is just a fallback, actual API routes are handled above.
			http.NotFound(w, r)
		} else {
			// For any other path, attempt to serve a static file.
			// This is a fallback for the http.Handle("/static/", ...)
			http.ServeFile(w, r, "./static"+r.URL.Path)
		}
	})

	log.Println("Server starting on :8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
