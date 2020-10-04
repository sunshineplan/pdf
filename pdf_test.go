package pdf

import (
	"bytes"
	"image"
	"io/ioutil"
	"testing"
)

func Test(t *testing.T) {
	b, err := ioutil.ReadFile("testdata/video-001.pdf")
	if err != nil {
		t.Error(err)
		return
	}
	if _, err := Decode(bytes.NewBuffer(b)); err != nil {
		t.Error("Failed to decode pdf", err)
	}
	if _, err := DecodeConfig(bytes.NewBuffer(b)); err != nil {
		t.Error("Failed to decode pdf config", err)
	}

	img, _, err := image.Decode(bytes.NewBuffer(b))
	if err != nil {
		t.Error("Failed to decode pdf", err)
	}
	if err := Encode(ioutil.Discard, []image.Image{img}, nil); err != nil {
		t.Error("Failed to encode pdf", err)
	}
}
