package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/aaronland/go-california-landscapes"
)

func main() {

	var destination string
	var verbose bool

	flag.StringVar(&destination, "destination", ".", "The destination folder where images should be written. Default is the current working directory.")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose (debug) logging.")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Fetch one or more posts from the California Landscapes bot and create a new JPEG image with EXIF data derived from the post.\n")
		fmt.Fprintf(os.Stderr, "Usage:\n\t %s [options] url(N) url(N)\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "If url(N) is \"-\" then the list of URLs to fetch is read from STDIN. Valid options are:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Debug("Verbose logging enabled")
	}

	ctx := context.Background()
	var root string

	if destination == "." {

		cwd, err := os.Getwd()

		if err != nil {
			log.Fatalf("Failed to determine current working directory, %v", err)
		}

		root = cwd
	} else {

		abs_path, err := filepath.Abs(destination)

		if err != nil {
			log.Fatalf("Failed to derive absolute path for '%s', %v", destination, err)
		}

		root = abs_path
	}

	uris := flag.Args()

	if len(uris) == 1 && uris[0] == "-" {

		uris = make([]string, 0)

		scanner := bufio.NewScanner(os.Stdin)

		for scanner.Scan() {
			uris = append(uris, strings.TrimSpace(scanner.Text()))
		}

		err := scanner.Err()

		if err != nil {
			log.Fatalf("Failed to scan STDIN, %v", err)
		}
	}

	for _, url := range uris {

		logger := slog.Default()
		logger = logger.With("url", url)

		base := filepath.Base(url)
		im_fname := fmt.Sprintf("%s.jpg", base)
		im_path := filepath.Join(root, im_fname)

		logger = logger.With("path", im_path)

		im_wr, err := os.OpenFile(im_path, os.O_RDWR|os.O_CREATE, 0644)

		if err != nil {
			log.Fatalf("Failed to open %s for writing, %v", im_path)
		}

		err = landscapes.DeriveImageFromPostURL(ctx, url, im_wr)

		if err != nil {
			// log.Fatalf("Failed to derive image from URL %s, %v", url, err)
			logger.Error("Failed to derive image, skipping", "error", err)
			im_wr.Close()
			os.Remove(im_path)
			continue
		}

		err = im_wr.Close()

		if err != nil {
			log.Fatalf("Failed to close %s after writing, %v", im_path, err)
		}

		logger.Info("Successfully fetched image")
	}
}
