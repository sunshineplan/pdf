package pdf

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"io/ioutil"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
)

// DefaultQuality is the default quality encoding parameter.
const DefaultQuality = 75

// Options are the encoding parameters. Quality ranges from 1 to 100 inclusive, higher is better.
type Options struct {
	Quality int
}

func decode(r io.Reader) (io.Reader, error) {
	conf := pdfcpu.NewDefaultConfiguration()
	conf.ValidationMode = pdfcpu.ValidationNone
	b, err := ioutil.ReadAll(r)
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
	if ctx.PageCount > 0 {
		imgs, err := ctx.ExtractPageImages(1)
		if err != nil || len(imgs) == 0 {
			return nil, fmt.Errorf("extract page images error: %v", err)
		}
		return imgs[0], nil
	}
	return nil, err
}

// Decode reads a PDF file from r and returns first image as an image.Image.
func Decode(r io.Reader) (image.Image, error) {
	pr, err := decode(r)
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(pr)
	return img, err
}

// DecodeConfig returns the color model and dimensions of a PDF first image without
// decoding the entire image.
func DecodeConfig(r io.Reader) (cfg image.Config, err error) {
	var pr io.Reader
	pr, err = decode(r)
	if err != nil {
		return
	}
	cfg, _, err = image.DecodeConfig(pr)
	return
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
