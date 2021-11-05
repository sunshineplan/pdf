package pdf

import (
	"bytes"
	"image"
	"io"
	"os"
	"testing"
)

func Test(t *testing.T) {
	b, err := os.ReadFile("testdata/video-001.pdf")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := image.Decode(bytes.NewBuffer(b)); err != nil {
		t.Error("Failed to decode pdf", err)
	}
	if _, _, err := image.DecodeConfig(bytes.NewBuffer(b)); err != nil {
		t.Error("Failed to decode pdf config", err)
	}

	imgs, err := DecodeAll(bytes.NewBuffer(b))
	if err != nil {
		t.Fatal("Failed to decode pdf", err)
	}
	if err := Encode(io.Discard, imgs, nil); err != nil {
		t.Fatal("Failed to encode pdf", err)
	}
}

func TestDecodeAll(t *testing.T) {
	f, err := os.Open("testdata/testImage.pdf")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	imgs, err := DecodeAll(f)
	if err != nil {
		t.Fatal(err)
	}
	if l := len(imgs); l != 2 {
		t.Errorf("want 2 images, got %d", l)
	}
}

func TestEncodeFile(t *testing.T) {
	if err := EncodeFile("test.tmp", []string{"testdata/video-001.pdf"}, nil); err != nil {
		t.Fatal(err)
	}
	if err := EncodeFile("test.tmp", []string{"testdata/testImage.pdf"}, nil); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open("test.tmp")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	imgs, err := DecodeAll(f)
	if err != nil {
		t.Fatal(err)
	}
	if l := len(imgs); l != 2 {
		t.Errorf("want 2 images, got %d", l)
	}
	f.Close()
	os.Remove("test.tmp")
}
