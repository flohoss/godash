package media

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
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
