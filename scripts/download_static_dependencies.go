package main

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	const url = "https://unpkg.com/htmx.org@1.9.10/dist/htmx.min.js"
	const dest = "static/js/htmx.min.js"

	err := os.MkdirAll(filepath.Dir(dest), 0755)
	if err != nil {
		panic(err)
	}

	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	out, err := os.Create(dest)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		panic(err)
	}
}
