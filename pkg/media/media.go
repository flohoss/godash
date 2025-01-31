package media

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
)

func DownloadSelfHostedIcon(url, title, filePath string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to get icon: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get icon, status: %d, url: %s", resp.StatusCode, url)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read icon: %w", err)
	}
	err = os.WriteFile(filePath, data, fs.FileMode(0640))
	if err != nil {
		return "", fmt.Errorf("failed to write icon: %w", err)
	}
	return filePath, nil
}
