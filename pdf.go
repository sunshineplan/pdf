package pdf

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"os"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
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
	conf := model.NewDefaultConfiguration()
	conf.ValidationMode = model.ValidationNone

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
		imgs, err := pdfcpu.ExtractPageImages(ctx, p, false)
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
		return tiff.Decode(bytes.NewBuffer(b))
	}

	return img, err
}

// Decode decodes a PDF file from r and returns first image as image.Image.
func Decode(r io.Reader) (image.Image, error) {
	pr, err := decodePDF(r)
	if err != nil {
		return nil, err
	}

	return decodeImage(pr[0])
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

// EncodeReader writes r to w.
func EncodeReader(w io.Writer, r []io.Reader) error {
	return api.ImportImages(nil, w, r, nil, nil)
}

// Append appends images to w.
func Append(rs io.ReadSeeker, w io.Writer, imgs []io.Reader) error {
	return api.ImportImages(rs, w, imgs, nil, nil)
}

func processImage(imgs []image.Image, o *Options) ([]io.Reader, error) {
	quality := DefaultQuality
	if o != nil {
		quality = o.Quality
		if quality < 1 {
			quality = 1
		} else if quality > 100 {
			quality = 100
		}
	}

	var r []io.Reader
	for _, i := range imgs {
		var buf bytes.Buffer
		if err := jpeg.Encode(&buf, i, &jpeg.Options{Quality: quality}); err != nil {
			return nil, err
		}
		r = append(r, &buf)
	}

	return r, nil
}

// Encode writes images to w.
func Encode(w io.Writer, imgs []image.Image, o *Options) error {
	r, err := processImage(imgs, o)
	if err != nil {
		return err
	}

	return EncodeReader(w, r)
}

func fileExists(filename string) bool {
	if _, err := os.Stat(filename); err == nil {
		return true
	}
	return false
}

// EncodeFile appends images to outFile which will be created if necessary.
func EncodeFile(outFile string, imgFiles []string, o *Options) (err error) {
	var f1, f2 *os.File
	rs := io.ReadSeeker(nil)
	tmpFile := outFile
	if fileExists(outFile) {
		if f1, err = os.Open(outFile); err != nil {
			return err
		}
		rs = f1
		tmpFile += ".tmp"
	}

	var imgs []image.Image
	for _, i := range imgFiles {
		f, err := os.Open(i)
		if err != nil {
			return err
		}
		defer f.Close()
		img, err := decodeImage(f)
		if err != nil {
			return err
		}
		imgs = append(imgs, img)
	}

	r, err := processImage(imgs, o)
	if err != nil {
		return err
	}

	if f2, err = os.Create(tmpFile); err != nil {
		return err
	}

	defer func() {
		if err != nil {
			f2.Close()
			if f1 != nil {
				f1.Close()
				os.Remove(tmpFile)
			}
			return
		}
		if err = f2.Close(); err != nil {
			return
		}
		if f1 != nil {
			if err = f1.Close(); err != nil {
				return
			}
			if err = os.Rename(tmpFile, outFile); err != nil {
				return
			}
		}
	}()

	return Append(rs, f2, r)
}

func init() {
	image.RegisterFormat("pdf", "%PDF", Decode, DecodeConfig)
}
