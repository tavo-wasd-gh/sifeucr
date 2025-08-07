package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const (
	url = "https://git.tavo.one/tavo/axiom/raw/branch/main/res/css/axiom.min.css"
	destDir = "static/css"
)


func main() {
	// Extract the filename from the URL
	filename := filepath.Base(url)
	destPath := filepath.Join(destDir, filename)

	// Create the destination directory if it doesn't exist
	err := os.MkdirAll(destDir, os.ModePerm)
	if err != nil {
		fmt.Printf("Failed to create directory: %v\n", err)
		os.Exit(1)
	}

	// Download the file
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Failed to download file: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Failed to fetch file: HTTP %d\n", resp.StatusCode)
		os.Exit(1)
	}

	// Create the output file
	out, err := os.Create(destPath)
	if err != nil {
		fmt.Printf("Failed to create file: %v\n", err)
		os.Exit(1)
	}
	defer out.Close()

	// Copy the response body to the file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Printf("Failed to save file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Downloaded '%s' to '%s'\n", url, destPath)
}
