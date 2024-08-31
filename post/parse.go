package post

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/anaskhan96/soup"
)

var re_coords = regexp.MustCompile(`(?:(\-?\d+(?:\.\d+)?), (\-?\d+(?:\.\d+)?))`)
var re_alt = regexp.MustCompile(`(\d+)\sft`)

type Image struct {
	PostURL       string
	ImageURL      string
	Image         image.Image
	Latitude      float64
	Longitude     float64
	Altitude      float64
	PublishedTime time.Time
}

func ParseURL(ctx context.Context, url string) (*Image, error) {

	logger := slog.Default()
	logger = logger.With("url", url)

	post_body, err := fetch(ctx, url)

	if err != nil {
		return nil, fmt.Errorf("Failed to fetch post, %w", err)
	}

	doc := soup.HTMLParse(string(post_body))

	meta_els := doc.FindAll("meta")

	// <meta content="https://botsin.space/@CALandscapeBot/113028507763072851" property="og:url">
	var post_url string

	// <meta content="https://files.botsin.space/media_attachments/files/113/028/507/654/450/155/original/8fc23194efc42460.jpeg" property="og:image">
	var image_url string

	// <meta content="2024-08-26T13:15:58Z" property="og:published_time">
	var pubtime time.Time

	// <meta content="Attached: 1 image
	// Mt Lukens 1, LosAngeles County, CA
	// 🗺34.2687, -118.2369 🧭141° ⛰5016 ft
	// https://ops.alertcalifornia.org/cam-console/2181" property="og:description">
	var latitude float64
	var longitude float64
	var altitude float64

	for _, el := range meta_els {

		attrs := el.Attrs()

		prop, exists := attrs["property"]

		if !exists {
			continue
		}

		switch prop {
		case "og:url":
			post_url = attrs["content"]
		case "og:image":
			image_url = attrs["content"]
		case "og:published_time":

			str_t := attrs["content"]

			t, err := time.Parse(time.RFC3339, str_t)

			if err != nil {
				return nil, fmt.Errorf("Failed to parse time for %s (%s), %w", url, attrs["content"], err)
			}

			pubtime = t

		case "og:description":

			lat, lon, alt, err := parse_description(attrs["content"])

			if err != nil {
				return nil, fmt.Errorf("Failed to parse description for %s (%s), %w", url, attrs["content"], err)
			}

			latitude = lat
			longitude = lon
			altitude = alt

		default:
			logger.Debug("Unknown or unsupported property", "property", prop)
		}

		if image_url != "" && post_url != "" && latitude != 0.0 && longitude != 0.0 && altitude != 0.0 && !pubtime.IsZero() {
			break
		}
	}

	if image_url == "" {
		return nil, fmt.Errorf("Failed to derive image URL")
	}

	if post_url == "" {
		return nil, fmt.Errorf("Failed to derive post URL")
	}

	if latitude == 0.0 {
		return nil, fmt.Errorf("Failed to derive latitude")
	}

	if longitude == 0.0 {
		return nil, fmt.Errorf("Failed to derive longitude")
	}

	if altitude == 0.0 {
		return nil, fmt.Errorf("Failed to derive altitude")
	}

	if pubtime.IsZero() {
		return nil, fmt.Errorf("Failed to derive published time")
	}

	im_body, err := fetch(ctx, image_url)

	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve image URL %s, %w", image_url, err)
	}

	im_r := bytes.NewReader(im_body)

	im, _, err := image.Decode(im_r)

	if err != nil {
		return nil, fmt.Errorf("Failed to decode image %s, %w", image_url, err)
	}

	img := &Image{
		PostURL:       post_url,
		ImageURL:      image_url,
		Latitude:      latitude,
		Longitude:     longitude,
		Altitude:      altitude,
		Image:         im,
		PublishedTime: pubtime,
	}

	return img, nil
}

func fetch(ctx context.Context, url string) ([]byte, error) {

	logger := slog.Default()
	logger = logger.With("url", url)

	logger.Debug("Fetch URL")

	cl := http.Client{}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)

	if err != nil {
		return nil, fmt.Errorf("Failed to create new request for %s, %w", url, err)
	}

	rsp, err := cl.Do(req)

	if err != nil {
		return nil, fmt.Errorf("Failed to request %s, %w", url, err)
	}

	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Request for %s failed, %d (%s)", url, rsp.StatusCode, rsp.Status)
	}

	body, err := io.ReadAll(rsp.Body)

	if err != nil {
		return nil, fmt.Errorf("Failed to read body for %s, %w", url, err)
	}

	return body, nil
}

func parse_description(body string) (float64, float64, float64, error) {

	logger := slog.Default()

	if !re_coords.MatchString(body) {
		return 0, 0, 0, fmt.Errorf("Description failed coordinates pattern match")
	}

	if !re_alt.MatchString(body) {
		return 0, 0, 0, fmt.Errorf("Description failed altitude pattern match")
	}

	coords_m := re_coords.FindStringSubmatch(body)
	alt_m := re_alt.FindStringSubmatch(body)

	str_lat := coords_m[1]
	str_lon := coords_m[2]
	str_alt := alt_m[1]

	logger.Debug("Parse description string values", "latitude", str_lat, "longitude", str_lon, "altitude", str_alt)

	lat, err := strconv.ParseFloat(str_lat, 64)

	if err != nil {
		return 0, 0, 0, fmt.Errorf("Failed to parse latitude (%s), %w", str_lat, err)
	}

	lon, err := strconv.ParseFloat(str_lon, 64)

	if err != nil {
		return 0, 0, 0, fmt.Errorf("Failed to parse longitude (%s), %w", str_lon, err)
	}

	alt, err := strconv.ParseFloat(str_alt, 64)

	if err != nil {
		return 0, 0, 0, fmt.Errorf("Failed to parse altitude (%s), %w", str_alt, err)
	}

	return lat, lon, alt, err
}
