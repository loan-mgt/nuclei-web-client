package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type scanStatus struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

var scanStatuses sync.Map

func cleanURL(url string) string {
	url = strings.TrimSpace(url)
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")
	return strings.TrimSuffix(url, "/")
}

func hashString(input string) string {
	h := sha1.New()
	h.Write([]byte(input))
	return hex.EncodeToString(h.Sum(nil))
}

func createFiles(hash, url string) error {
	dirPath := filepath.Join("./data", hash)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return err
	}

	urlsFilePath := filepath.Join(dirPath, "urls.txt")
	initFilePath := filepath.Join(dirPath, "INIT.txt")

	if err := os.WriteFile(urlsFilePath, []byte(url), 0644); err != nil {
		return err
	}

	initContent := []byte(time.Now().Format(time.UnixDate))
	return os.WriteFile(initFilePath, initContent, 0644)
}

func scanHandler(w http.ResponseWriter, r *http.Request) {
	target := r.FormValue("url")
	cleanedURL := cleanURL(target)
	hash := hashString(cleanedURL)

	scanStatuses.Store(hash, scanStatus{Status: "preparing", Message: "Creating necessary files..."})

	if err := createFiles(hash, cleanedURL); err != nil {
		http.Error(w, "Failed to prepare files", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/"+hash, http.StatusSeeOther)
	go runScan(hash)
}

func runScan(hash string) {
	dirPath := filepath.Join("./data", hash)
	filepath.Join(dirPath, "results.jsonl")

	// Update status to ongoing
	scanStatuses.Store(hash, scanStatus{Status: "ongoing", Message: "Scan in progress..."})

	cmd := exec.Command("docker", "run", "--name", hash, "--rm", "-v", dirPath+":/app/", "projectdiscovery/nuclei", "-l", "/app/urls.txt", "-jsonl", "/app/results.jsonl")
	if err := cmd.Start(); err != nil {
		scanStatuses.Store(hash, scanStatus{Status: "error", Message: "Failed to start scan."})
		return
	}

	cmd.Wait()

	// Finalize status
	scanStatuses.Store(hash, scanStatus{Status: "finished", Message: "Scan completed. Results are ready."})
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["hash"]

	status, _ := scanStatuses.LoadOrStore(hash, scanStatus{Status: "unknown", Message: "Scan not found."})
	json.NewEncoder(w).Encode(status)
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade WebSocket:", err)
		return
	}
	defer conn.Close()

	vars := mux.Vars(r)
	hash := vars["hash"]

	for {
		status, _ := scanStatuses.Load(hash)
		if err := conn.WriteJSON(status); err != nil {
			log.Println("Failed to send WebSocket message:", err)
			break
		}
		time.Sleep(5 * time.Second)
	}
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/scan", scanHandler).Methods("POST")
	r.HandleFunc("/status/{hash}", statusHandler).Methods("GET")
	r.HandleFunc("/ws/{hash}", wsHandler)

	staticDir := "./public"
	r.PathPrefix("/").Handler(http.FileServer(http.Dir(staticDir)))

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
