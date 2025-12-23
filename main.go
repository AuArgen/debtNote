package main

import (
	"debtNote/database"
	"debtNote/handlers"
	"embed"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
)

//go:embed static/*
var staticFiles embed.FS

func main() {
	// Initialize database
	database.InitDB()
	defer database.DB.Close()

	// Create uploads directory if it doesn't exist
	if _, err := os.Stat("uploads"); os.IsNotExist(err) {
		os.Mkdir("uploads", 0755)
	}

	// Get the static directory from the embedded file system
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatal(err)
	}

	// Serve static files (Embedded)
	fileServer := http.FileServer(http.FS(staticFS))
	http.Handle("/static/", http.StripPrefix("/static/", fileServer))

	// Serve uploaded files (Local file system)
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	// API routes
	http.HandleFunc("/api/clients", handlers.GetClientsHandler)
	http.HandleFunc("/api/clients/search", handlers.SearchClientsHandler)
	http.HandleFunc("/api/debts", handlers.GetDebtsHandler)
	http.HandleFunc("/api/debts/add", handlers.AddDebtHandler)
	http.HandleFunc("/api/debts/pay", handlers.MakePaymentHandler)
	http.HandleFunc("/api/debts/payments", handlers.GetDebtPaymentsHandler)
	http.HandleFunc("/api/debts/delete", handlers.DeleteDebtHandler) // New handler

	// Handle SPA (Single Page Application) routing
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/uploads/") {
			http.NotFound(w, r)
			return
		}
		serveIndex(w, staticFS)
	})

	log.Println("Server starting on :8080...")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func serveIndex(w http.ResponseWriter, staticFS fs.FS) {
	indexFile, err := staticFS.Open("index.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer indexFile.Close()

	content, err := io.ReadAll(indexFile)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(content)
}
