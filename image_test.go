package landscapes

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/aaronland/go-california-landscapes/post"
	"github.com/dsoprea/go-exif/v3"
	"github.com/dsoprea/go-exif/v3/common"
)

func TestDeriveImageFromPostURL(t *testing.T) {

	slog.SetLogLoggerLevel(slog.LevelDebug)

	ctx := context.Background()

	tests := map[string]*post.Image{
		"https://botsin.space/users/CALandscapeBot/statuses/113028507763072851": &post.Image{
			PostURL:   "https://botsin.space/@CALandscapeBot/113028507763072851",
			ImageURL:  "https://files.botsin.space/media_attachments/files/113/028/507/654/450/155/original/8fc23194efc42460.jpeg",
			Latitude:  34.2687,
			Longitude: -118.2369,
			// To do: figure out how to include PublishedTime here so it can be compared to DateTimeOriginal below
		},
	}

	for url, expected := range tests {

		logger := slog.Default()
		logger = logger.With("url", url)

		wr, err := os.CreateTemp("", "example.*.jpg")

		if err != nil {
			t.Fatalf("Failed to create temp file for %s, %v", url, err)
		}

		logger = logger.With("temp file", wr.Name())

		defer func() {

			logger.Debug("Remove temp file")
			err := os.Remove(wr.Name())

			if err != nil {
				logger.Error("Failed to remove temp file", "error", err)
			}
		}()

		logger.Debug("Derive image from post")

		err = DeriveImageFromPostURL(ctx, url, wr)

		if err != nil {
			t.Fatalf("Failed to derive image from post %s, %v", url, err)
		}

		err = wr.Close()

		if err != nil {
			t.Fatalf("Failed to close temp file writer for %s, %v", url, err)
		}

		logger.Debug("Extract EXIF data from new image")

		raw_exif, err := exif.SearchFileAndExtractExif(wr.Name())

		if err != nil {
			t.Fatalf("Failed to extract EXIF from temp file for %s, %v", url, err)
		}

		im, err := exifcommon.NewIfdMappingWithStandard()

		if err != nil {
			t.Fatalf("Failed to create IFD mapping for %s, %v", url, err)
		}

		ti := exif.NewTagIndex()

		_, index, err := exif.Collect(im, ti, raw_exif)

		if err != nil {
			t.Fatalf("Failed to collect EXIF data for %s, %v", url, err)
		}

		ifd, err := index.RootIfd.ChildWithIfdPath(exifcommon.IfdGpsInfoStandardIfdIdentity)

		if err != nil {
			t.Fatalf("Failed to derive IFD child for %s, %v", url, err)
		}

		gi, err := ifd.GpsInfo()

		if err != nil {
			t.Fatalf("Failed to derive GPS info for %s, %v", url, err)
		}

		lat := gi.Latitude.Decimal()
		lon := gi.Longitude.Decimal()

		logger.Debug("GPS data", "latitude", lat, "longitude", lon)

		// See the way only the first three decimal points are being compared? Computers, amirite...

		if fmt.Sprintf("%.03f", lat) != fmt.Sprintf("%.03f", expected.Latitude) {
			t.Fatalf("Unexpected latitude for %s, %.03f (expected %.03f)", url, lat, expected.Latitude)
			// logger.Warn("Unexpected latitude in EXIF data", "latitude", lat, "expected", expected.Latitude)
		}

		if fmt.Sprintf("%.03f", lon) != fmt.Sprintf("%.03f", expected.Longitude) {
			t.Fatalf("Unexpected longitude for %s, %.03f (expected %.03f)", url, lon, expected.Longitude)
			// logger.Warn("Unexpected longitude in EXIF data", "longitude", lon, "expected", expected.Longitude)
		}

		rootIfd := index.RootIfd

		results, err := rootIfd.FindTagWithName("DateTimeOriginal")

		if err != nil {
			t.Fatalf("Failed to find tag for DateTimeOriginal for %s, %v", url, err)
		}

		if len(results) != 1 {
			t.Fatalf("Multiple results for tag DateTimeOriginal for %s, %v", url, err)
		}

		v, err := results[0].Value()

		if err != nil {
			t.Fatalf("Failed to derive value for DateTimeOriginal tag for %s, %v", url, err)
		}

		logger.Debug("DateTimeOriginal data", "value", v)

	}

}
