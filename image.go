package landscapes

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"image/jpeg"
	"io"

	"github.com/aaronland/go-california-landscapes/post"
	"github.com/sfomuseum/go-exif-update"
)

func DeriveImageFromPostURL(ctx context.Context, url string, wr io.Writer) error {

	img, err := post.ParseURL(ctx, url)

	if err != nil {
		return fmt.Errorf("Failed to parse post URL, %w", err)
	}

	var buf bytes.Buffer
	im_wr := bufio.NewWriter(&buf)

	err = jpeg.Encode(im_wr, img.Image, &jpeg.Options{Quality: 100})

	if err != nil {
		return fmt.Errorf("Failed to encode post image, %w", err)
	}

	im_wr.Flush()

	im_r := bytes.NewReader(buf.Bytes())

	alt_m := int(img.Altitude * 0.30479999999999996)
	str_alt := fmt.Sprintf("%d/1", alt_m)

	exif_props := map[string]interface{}{
		"X-Latitude":  img.Latitude,
		"X-Longitude": img.Longitude,
		"GPSAltitude": str_alt,
		// Above sea level. I can not figure out what dsoprea/go-exif
		// needs this value to be...
		// "GPSAltitudeRef": "0",
		"DateTimeOriginal": img.PublishedTime.Format("2006:01:02 15:04:05"),
		"ImageID":          img.PostURL,
		"ImageUniqueID":    img.ImageURL,
	}

	return update.PrepareAndUpdateExif(im_r, wr, exif_props)
}
