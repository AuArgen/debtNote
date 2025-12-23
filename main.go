package main

import (
	"debtNote/database"
	"debtNote/handlers"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
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
	http.HandleFunc("/api/debts/delete", handlers.DeleteDebtHandler)

	// Handle SPA (Single Page Application) routing
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/uploads/") {
			http.NotFound(w, r)
			return
		}
		serveIndex(w, staticFS)
	})

	// Server address
	addr := ":8080"
	url := "http://localhost" + addr

	// Print clickable link
	fmt.Println("---------------------------------------------------------")
	fmt.Println(" Программа ийгиликтүү ишке кирди!")
	fmt.Println(" Программаны ачуу үчүн төмөнкү ссылканы басыңыз:")
	fmt.Printf(" \n %s \n\n", url)
	fmt.Println("---------------------------------------------------------")

	// Open browser automatically in a goroutine
	go func() {
		// Give the server a second to start
		time.Sleep(1 * time.Second)
		openBrowser(url)
	}()

	// Fix: Use '=' instead of ':=' because err is already declared
	err = http.ListenAndServe(addr, nil)
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

// openBrowser opens the specified URL in the default browser of the user.
func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	if err != nil {
		log.Printf("Браузерди ачууда ката кетти: %v\nСураныч, ссылканы кол менен ачыңыз: %s", err, url)
	}
}
