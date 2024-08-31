package post

import (
	"context"
	"log/slog"
	"testing"
)

func TestParseURL(t *testing.T) {

	slog.SetLogLoggerLevel(slog.LevelDebug)

	ctx := context.Background()

	tests := map[string]*Image{
		"https://botsin.space/users/CALandscapeBot/statuses/113028507763072851": &Image{
			PostURL:   "https://botsin.space/@CALandscapeBot/113028507763072851",
			ImageURL:  "https://files.botsin.space/media_attachments/files/113/028/507/654/450/155/original/8fc23194efc42460.jpeg",
			Latitude:  34.2687,
			Longitude: -118.2369,
		},
	}

	for url, expected := range tests {

		img, err := ParseURL(ctx, url)

		if err != nil {
			t.Fatalf("Failed to parse %s, %v", url, err)
		}

		if img.PostURL != expected.PostURL {
			t.Fatalf("Unexpected post URL for %s (%s)", url, img.PostURL)
		}

		if img.ImageURL != expected.ImageURL {
			t.Fatalf("Unexpected image URL for %s (%s)", url, img.ImageURL)
		}

		if img.Latitude != expected.Latitude {
			t.Fatalf("Unexpected latitude for %s (%f)", url, img.Latitude)
		}

		if img.Longitude != expected.Longitude {
			t.Fatalf("Unexpected longitude for %s (%f)", url, img.Longitude)
		}

	}

}

func TestParseDescription(t *testing.T) {

	slog.SetLogLoggerLevel(slog.LevelDebug)

	tests := map[string][3]float64{
		`Attached: 1 image

Mt Lukens 1, LosAngeles County, CA
🗺34.2687, -118.2369 🧭141° ⛰5016 ft
https://ops.alertcalifornia.org/cam-console/2181`: [3]float64{34.2687, -118.2369, 5016},
	}

	for text, coords := range tests {

		lat, lon, alt, err := parse_description(text)

		if err != nil {
			t.Fatalf("Failed to parse description, %v (%s)", err, text)
		}

		if lat != coords[0] {
			t.Fatalf("Unexpected latitude for description: '%f' (%s)", lat, text)
		}

		if lon != coords[1] {
			t.Fatalf("Unexpected longitude for description: '%f' (%s)", lon, text)
		}

		if alt != coords[2] {
			t.Fatalf("Unexpected altitude for description: '%f' (%s)", alt, text)
		}

	}
}
