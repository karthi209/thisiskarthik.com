package imageprivacy

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"
)

func TestRasterMetadataRemovedWithoutChangingPixels(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	for _, ext := range []string{".png", ".jpg"} {
		var encoded bytes.Buffer
		if ext == ".png" {
			if err := png.Encode(&encoded, img); err != nil {
				t.Fatal(err)
			}
		} else {
			if err := jpeg.Encode(&encoded, img, nil); err != nil {
				t.Fatal(err)
			}
		}
		original := encoded.Bytes()
		var annotated []byte
		if ext == ".png" {
			chunk := make([]byte, 12+len("GPS location"))
			binary.BigEndian.PutUint32(chunk[:4], uint32(len("GPS location")))
			copy(chunk[4:8], "eXIf")
			copy(chunk[8:], "GPS location")
			annotated = append(append(append([]byte{}, original[:8]...), chunk...), original[8:]...)
		} else {
			segment := []byte{0xff, 0xe1, 0, 14}
			segment = append(segment, []byte("GPS location")...)
			annotated = append(append(append([]byte{}, original[:2]...), segment...), original[2:]...)
		}
		clean, err := Strip(annotated, ext)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(clean, original) {
			t.Fatalf("%s compressed pixels changed", ext)
		}
		if _, _, err := image.Decode(bytes.NewReader(clean)); err != nil {
			t.Fatal(err)
		}
		if !bytes.Contains(annotated, []byte("GPS location")) {
			t.Fatal("source modified")
		}
	}
}

func TestWebPMetadataAndFlags(t *testing.T) {
	chunk := func(tag string, payload []byte) []byte {
		out := make([]byte, 8)
		copy(out, tag)
		binary.LittleEndian.PutUint32(out[4:], uint32(len(payload)))
		out = append(out, payload...)
		if len(payload)%2 != 0 {
			out = append(out, 0)
		}
		return out
	}
	raw := []byte("RIFF\x00\x00\x00\x00WEBP")
	flags := make([]byte, 10)
	flags[0] = 0x3c
	for _, c := range [][]byte{chunk("VP8X", flags), chunk("EXIF", []byte("GPS")), chunk("XMP ", []byte("account")), chunk("ICCP", []byte("color")), chunk("VP8 ", []byte("pixels"))} {
		raw = append(raw, c...)
	}
	binary.LittleEndian.PutUint32(raw[4:8], uint32(len(raw)-8))
	clean, err := Strip(raw, ".webp")
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(clean, []byte("GPS")) || bytes.Contains(clean, []byte("account")) || clean[20] != 0x30 || !bytes.Contains(clean, []byte("pixels")) || !bytes.Contains(clean, []byte("color")) {
		t.Fatal("incorrect metadata removal")
	}
	if int(binary.LittleEndian.Uint32(clean[4:8])) != len(clean)-8 {
		t.Fatal("incorrect RIFF size")
	}
	again, err := Strip(clean, ".webp")
	if err != nil || !bytes.Equal(clean, again) {
		t.Fatal("stripping is not idempotent")
	}
}

func TestMalformedRasterRejected(t *testing.T) {
	for _, ext := range []string{".png", ".webp", ".jpg"} {
		if _, err := Strip([]byte("broken"), ext); err == nil {
			t.Errorf("accepted broken %s", ext)
		}
	}
}
