package utils

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func CleanURL(url string) string {
	url = strings.TrimSpace(url)
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")
	return strings.TrimSuffix(url, "/")
}

func HashString(input string) string {
	h := sha1.New()
	h.Write([]byte(input))
	return hex.EncodeToString(h.Sum(nil))
}

func CreateFiles(hash, url string) error {
	dirPath := filepath.Join("./data", hash)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return err
	}

	urlsFilePath := filepath.Join(dirPath, "urls.txt")
	initFilePath := filepath.Join(dirPath, "INIT.txt")

	// Write the cleaned URL to urls.txt
	if err := os.WriteFile(urlsFilePath, []byte(url), 0644); err != nil {
		return err
	}

	// Write the current UNIX timestamp to INIT.txt
	unixTimestamp := []byte(fmt.Sprintf("%d", time.Now().Unix()))
	return os.WriteFile(initFilePath, unixTimestamp, 0644)
}
