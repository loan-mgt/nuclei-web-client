package scan

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"nuclei/web-client/internal/models"
)

func RunScan(hash string) {
	dirPath := filepath.Join("./data", hash)
	resultsFilePath := filepath.Join(dirPath, "results.jsonl")

	// Update status to ongoing
	models.ScanStatuses.Store(hash, models.ScanStatus{
		Status:  "ongoing",
		Message: "Scan in progress...",
		Output:  "",
	})

	// Execute the Nuclei scan
	cmd := exec.Command("nuclei", "-l", filepath.Join(dirPath, "urls.txt"), "-jsonl", resultsFilePath)

	// Create a pipe to capture the command's output
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		models.ScanStatuses.Store(hash, models.ScanStatus{
			Status:  "error",
			Message: "Failed to capture command output: " + err.Error(),
			Output:  "",
		})
		return
	}

	if err := cmd.Start(); err != nil {
		models.ScanStatuses.Store(hash, models.ScanStatus{
			Status:  "error",
			Message: "Failed to start scan: " + err.Error(),
			Output:  "",
		})
		return
	}

	// Start a goroutine to read the output of the command and periodically update the status
	if f, ok := stdout.(*os.File); ok {
		go captureCommandOutput(hash, f)
	} else {
		models.ScanStatuses.Store(hash, models.ScanStatus{
			Status:  "error",
			Message: "Failed to capture command output",
			Output:  "",
		})
		return
	}

	// Wait for the command to complete
	cmd.Wait()

	// Finalize status when scan is finished
	models.ScanStatuses.Store(hash, models.ScanStatus{
		Status:  "finished",
		Message: "Scan completed. Results are ready.",
		Output:  "Scan completed.",
	})
}

// captureCommandOutput reads the command's output and updates the status every 5 seconds
func captureCommandOutput(hash string, stdout *os.File) {
	scanner := bufio.NewScanner(stdout)
	var output strings.Builder

	// Read the output line by line and update status every 5 seconds
	for scanner.Scan() {
		line := scanner.Text()
		output.WriteString(line + "\n")

		// Update status every 5 seconds
		models.ScanStatuses.Store(hash, models.ScanStatus{
			Status:  "ongoing",
			Message: "Scan in progress...",
			Output:  output.String(),
		})

		time.Sleep(5 * time.Second) // Sleep for 5 seconds before updating again
	}

	if err := scanner.Err(); err != nil {
		models.ScanStatuses.Store(hash, models.ScanStatus{
			Status:  "error",
			Message: "Error reading output: " + err.Error(),
			Output:  output.String(),
		})
	}
}
