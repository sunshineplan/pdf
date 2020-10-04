package pdf

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestDecode(t *testing.T) {
	b, err := ioutil.ReadFile("video-001.pdf")
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
}
