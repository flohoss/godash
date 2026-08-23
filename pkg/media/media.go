package media

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"strings"
	"time"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

func DownloadSelfHostedIcon(url, title, filePath string) (string, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to get icon: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get icon, status: %d, url: %s", resp.StatusCode, url)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		return "", fmt.Errorf("failed to read icon: %w", err)
	}
	if !isIconContentType(resp.Header.Get("Content-Type"), data) {
		return "", fmt.Errorf("downloaded icon is not an image: %s", http.DetectContentType(data))
	}
	tmpPath := filePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, fs.FileMode(0640)); err != nil {
		return "", fmt.Errorf("failed to write icon: %w", err)
	}
	if err := os.Rename(tmpPath, filePath); err != nil {
		os.Remove(tmpPath)
		return "", fmt.Errorf("failed to move icon: %w", err)
	}
	return filePath, nil
}

func isIconContentType(header string, data []byte) bool {
	ct := strings.ToLower(strings.TrimSpace(strings.Split(header, ";")[0]))
	if strings.HasPrefix(ct, "image/") {
		return true
	}
	switch strings.Split(strings.ToLower(strings.TrimSpace(http.DetectContentType(data))), ";")[0] {
	case "image/svg+xml", "text/xml", "application/xml":
		return true
	}
	return false
}
