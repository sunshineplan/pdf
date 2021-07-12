package pdf

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/sunshineplan/tiff"
	_ "golang.org/x/image/webp"
)

// DefaultQuality is the default quality encoding parameter.
const DefaultQuality = 75

// Options are the encoding parameters. Quality ranges from 1 to 100 inclusive, higher is better.
type Options struct {
	Quality int
}

func decodePDF(r io.Reader) ([]io.Reader, error) {
	conf := pdfcpu.NewDefaultConfiguration()
	conf.ValidationMode = pdfcpu.ValidationNone

	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	ctx, err := pdfcpu.Read(bytes.NewReader(b), conf)
	if err != nil {
		return nil, err
	}

	if err := api.OptimizeContext(ctx); err != nil {
		return nil, err
	}
	if ctx.PageCount == 0 {
		return nil, fmt.Errorf("page count is zero")
	}

	var rs []io.Reader
	for p := 1; p <= ctx.PageCount; p++ {
		imgs, err := ctx.ExtractPageImages(p, false)
		if err != nil {
			return nil, err
		}

		for _, img := range imgs {
			if img.Reader != nil {
				rs = append(rs, img)
			}
		}
	}
	if len(rs) == 0 {
		return nil, fmt.Errorf("no image found")
	}

	return rs, nil
}

func decodeImage(r io.Reader) (image.Image, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	img, format, err := image.Decode(bytes.NewBuffer(b))
	if format == "tiff" && err != nil {
		img, err = tiff.Decode(bytes.NewBuffer(b))
	}

	return img, err
}

// Decode decodes a PDF file from r and returns first image as image.Image.
func Decode(r io.Reader) (image.Image, error) {
	pr, err := decodePDF(r)
	if err != nil {
		return nil, err
	}

	img, err := decodeImage(pr[0])

	return img, err
}

// DecodeAll decodes a PDF file from r and returns all images as image.Image.
func DecodeAll(r io.Reader) ([]image.Image, error) {
	pr, err := decodePDF(r)
	if err != nil {
		return nil, err
	}

	var imgs []image.Image
	for _, r := range pr {
		img, err := decodeImage(r)
		if err != nil {
			return nil, err
		}
		imgs = append(imgs, img)
	}

	return imgs, err
}

// DecodeConfig returns the color model and dimensions of a PDF first image without
// decoding the entire image.
func DecodeConfig(r io.Reader) (image.Config, error) {
	pr, err := decodePDF(r)
	if err != nil {
		return image.Config{}, err
	}
	cfg, _, err := image.DecodeConfig(pr[0])

	return cfg, err
}

// Encode writes images to w.
func Encode(w io.Writer, imgs []image.Image, o *Options) error {
	quality := DefaultQuality
	if o != nil {
		quality = o.Quality
		if quality < 1 {
			quality = 1
		} else if quality > 100 {
			quality = 100
		}
	}

	var rs []io.Reader
	for _, i := range imgs {
		var buf bytes.Buffer
		if err := jpeg.Encode(&buf, i, &jpeg.Options{Quality: quality}); err != nil {
			return err
		}
		rs = append(rs, &buf)
	}

	return api.ImportImages(nil, w, rs, nil, nil)
}

func init() {
	image.RegisterFormat("pdf", "%PDF", Decode, DecodeConfig)
}
