package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"nuclei/web-client/internal/models"
	"nuclei/web-client/internal/scan"
	"nuclei/web-client/internal/utils"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func ScanHandler(w http.ResponseWriter, r *http.Request) {
	target := r.FormValue("url")
	cleanedURL := utils.CleanURL(target)
	hash := utils.HashString(cleanedURL)

	models.ScanStatuses.Store(hash, models.ScanStatus{Status: "preparing", Message: "Creating necessary files..."})

	if err := utils.CreateFiles(hash, cleanedURL); err != nil {
		http.Error(w, "Failed to prepare files", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/"+hash, http.StatusSeeOther)
	go scan.RunScan(hash)
}

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["hash"]

	status, _ := models.ScanStatuses.LoadOrStore(hash, models.ScanStatus{Status: "unknown", Message: "Scan not found."})
	json.NewEncoder(w).Encode(status)
}

func WSHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade WebSocket:", err)
		return
	}
	defer conn.Close()

	vars := mux.Vars(r)
	hash := vars["hash"]

	for {
		status, _ := models.ScanStatuses.Load(hash)
		if err := conn.WriteJSON(status); err != nil {
			log.Println("Failed to send WebSocket message:", err)
			break
		}
		time.Sleep(5 * time.Second)
	}
}
