package main

import (
	"log"
	"net/http"

	"nuclei/web-client/internal/handlers"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()

	// Register routes
	r.HandleFunc("/scan", handlers.ScanHandler).Methods("POST")
	r.HandleFunc("/status/{hash}", handlers.StatusHandler).Methods("GET")
	r.HandleFunc("/ws/{hash}", handlers.WSHandler)

	// Serve static files
	staticDir := "./public"
	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir(staticDir))))

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
