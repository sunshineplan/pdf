package pdf

import (
	"os"
	"testing"
)

func TestDecode(t *testing.T) {
	f, err := os.Open("video-001.pdf")
	if err != nil {
		t.Error(err)
		return
	}
	if _, err := Decode(f); err != nil {
		t.Error("Failed to decode pdf")
	}
	if _, err := DecodeConfig(f); err != nil {
		t.Error("Failed to decode pdf config")
	}
}
