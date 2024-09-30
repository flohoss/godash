package media

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"regexp"
	"strings"
)

func DownloadSelfHostedIcon(ext, title, filePath string) ([]byte, error) {
	resp, err := http.Get("https://cdn.jsdelivr.net/gh/selfhst/icons/" + strings.TrimPrefix(ext, ".") + "/" + title)
	if err != nil {
		return nil, fmt.Errorf("failed to get icon: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get icon, status: %d", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read icon: %w", err)
	}
	data = replaceClassNames(data, strings.TrimSuffix(title, ext))
	data = insertWidthHeight(data)
	err = os.WriteFile(filePath, data, fs.FileMode(0640))
	if err != nil {
		return nil, fmt.Errorf("failed to write icon: %w", err)
	}
	return data, nil
}

func insertWidthHeight(svgContent []byte) []byte {
	classRegex := regexp.MustCompile(`(?:<svg(.+?)(width|height)=".+?")(.+?)(width|height)=".+?"|(<svg)`)
	newSVGContent := classRegex.ReplaceAllFunc(svgContent, func(match []byte) []byte {
		groups := classRegex.FindSubmatch(match)
		if len(groups) == 0 {
			return match
		}
		if string(match) == "<svg" {
			return []byte(`<svg width="2rem" height="2rem" `)
		} else {
			return []byte(fmt.Sprintf(`<svg%s%s="2rem"%s%s="2rem"`, groups[1], groups[2], groups[3], groups[4]))
		}
	})
	return newSVGContent
}

func replaceClassNames(svgContent []byte, title string) []byte {
	// Regular expression to match either class="st0" or .st0
	classRegex := regexp.MustCompile(`(class="|\.)([a-z]{2}\d)`)

	newSVGContent := classRegex.ReplaceAllFunc(svgContent, func(match []byte) []byte {
		groups := classRegex.FindSubmatch(match)
		if len(groups) == 0 {
			return match
		}
		group1 := string(groups[1])
		group2 := string(groups[2])
		if group1 == `class="` {
			return []byte(`class="` + title + "-" + group2)
		}
		if group1 == `.` {
			return []byte(`.` + title + "-" + group2)
		}
		return match
	})

	return newSVGContent
}
