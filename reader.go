package pdf

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"log"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

type Reader struct {
	ctx *model.Context
	p   int
}

func NewReader(rs io.ReadSeeker) (*Reader, error) {
	conf := model.NewDefaultConfiguration()
	conf.ValidationMode = model.ValidationNone
	ctx, err := pdfcpu.Read(rs, conf)
	if err != nil {
		return nil, err
	}
	if err := api.OptimizeContext(ctx); err != nil {
		return nil, err
	}
	if ctx.PageCount == 0 {
		return nil, fmt.Errorf("page count is zero")
	}
	return &Reader{ctx: ctx}, nil
}

func (r Reader) PageCount() int {
	return r.ctx.PageCount
}

func (r *Reader) Next() bool {
	r.p++
	return r.p <= r.ctx.PageCount
}

func (r *Reader) Extract() (rs []io.Reader, err error) {
	res, err := pdfcpu.ExtractPageImages(r.ctx, r.p, false)
	if err != nil {
		return
	}
	for _, v := range res {
		rs = append(rs, v)
	}
	return
}

func (r *Reader) ExtractPageImages(pageNr int) (imgs []image.Image, err error) {
	m, err := pdfcpu.ExtractPageImages(r.ctx, pageNr, false)
	if err != nil {
		return
	}
	for _, r := range m {
		var img image.Image
		img, _, err = image.Decode(r)
		if err != nil {
			return
		}
		imgs = append(imgs, img)
	}
	return
}

func (r *Reader) ExtractImages() ([]image.Image, error) {
	return r.ExtractPageImages(r.p)
}

func newReader(r io.Reader) (*Reader, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return NewReader(bytes.NewReader(b))
}

// Decode decodes a PDF file from r and returns first image as image.Image.
func Decode(r io.Reader) (image.Image, error) {
	reader, err := newReader(r)
	if err != nil {
		return nil, err
	}
	reader.Next()
	imgs, err := reader.ExtractImages()
	if err != nil {
		return nil, err
	}
	return imgs[0], nil
}

// DecodeAll decodes a PDF file from r and returns all images as image.Image.
func DecodeAll(r io.Reader) ([]image.Image, error) {
	reader, err := newReader(r)
	if err != nil {
		return nil, err
	}
	var imgs []image.Image
	for reader.Next() {
		res, err := reader.ExtractImages()
		if err != nil {
			log.Print(err)
			continue
		}
		imgs = append(imgs, res...)
	}
	return imgs, err
}

// DecodeConfig returns the color model and dimensions of a PDF first image without
// decoding the entire image.
func DecodeConfig(r io.Reader) (cfg image.Config, err error) {
	reader, err := newReader(r)
	if err != nil {
		return
	}
	reader.Next()
	rs, err := reader.Extract()
	if err != nil {
		return
	}
	cfg, _, err = image.DecodeConfig(rs[0])
	return
}

func init() {
	image.RegisterFormat("pdf", "%PDF", Decode, DecodeConfig)
}
