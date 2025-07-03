package main

import (
	"flag"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"github.com/tdewolff/minify/v2/json"
	"github.com/tdewolff/minify/v2/svg"
	"github.com/tdewolff/minify/v2/xml"
)

func main() {
	srcDir := flag.String("src", "", "Source directory to minify")
	dstDir := flag.String("dst", "", "Destination directory for minified files")
	flag.Parse()

	if *srcDir == "" || *dstDir == "" {
		fmt.Println("Usage: minifyfolder -src path/to/source -dst path/to/dest")
		os.Exit(1)
	}

	m := minify.New()
	m.AddFunc("text/html", html.Minify)
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("application/javascript", js.Minify)
	m.AddFunc("application/json", json.Minify)
	m.AddFunc("image/svg+xml", svg.Minify)
	m.AddFunc("text/xml", xml.Minify)

	err := filepath.WalkDir(*srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil // skip directories
		}

		ext := strings.ToLower(filepath.Ext(path))
		mime := mimeTypeFromExt(ext)
		if mime == "" {
			// unsupported file type, skip
			return nil
		}

		relPath, err := filepath.Rel(*srcDir, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(*dstDir, relPath)
		err = os.MkdirAll(filepath.Dir(dstPath), 0755)
		if err != nil {
			return err
		}

		err = minifyFile(m, mime, path, dstPath)
		if err != nil {
			return fmt.Errorf("failed to minify %s: %w", path, err)
		}

		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error during minify: %v\n", err)
		os.Exit(1)
	}
}

func mimeTypeFromExt(ext string) string {
	switch ext {
	case ".html", ".htm":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".json":
		return "application/json"
	case ".svg":
		return "image/svg+xml"
	case ".xml":
		return "text/xml"
	default:
		return ""
	}
}

func minifyFile(m *minify.M, mime, srcPath, dstPath string) error {
	data, err := ioutil.ReadFile(srcPath)
	if err != nil {
		return err
	}

	minified, err := m.Bytes(mime, data)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(dstPath, minified, 0644)
}
