package pdf

import (
	"bytes"
	"image"
	"io"
	"io/ioutil"

	"github.com/disintegration/imaging"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
)

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
		img, err := ctx.ExtractPageImages(1)
		if err != nil {
			return nil, err
		}
		return img[0], nil
	}
	return nil, err
}

// Decode reads a PDF file from r and returns first image as an image.Image.
func Decode(r io.Reader) (image.Image, error) {
	pr, err := decode(r)
	if err != nil {
		return nil, err
	}
	img, _, err := imaging.Decode(pr)
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
	cfg, _, err = imaging.DecodeConfig(pr)
	return
}

func init() {
	image.RegisterFormat("pdf", "PDF", Decode, DecodeConfig)
}
