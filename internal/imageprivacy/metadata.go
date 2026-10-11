// Package imageprivacy removes nonvisual metadata from published raster copies.
package imageprivacy

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
)

// Strip preserves compressed pixels and color profiles. Uploads should have their
// orientation normalized first, since EXIF orientation is intentionally removed.
func Strip(data []byte, ext string) ([]byte, error) {
	switch strings.ToLower(ext) {
	case ".webp":
		return stripWebP(data)
	case ".png":
		return stripPNG(data)
	case ".jpg", ".jpeg":
		return stripJPEG(data)
	default:
		return data, nil
	}
}

func stripWebP(data []byte) ([]byte, error) {
	if len(data) < 12 || string(data[:4]) != "RIFF" || string(data[8:12]) != "WEBP" || uint64(binary.LittleEndian.Uint32(data[4:8]))+8 != uint64(len(data)) {
		return nil, fmt.Errorf("invalid WebP container")
	}
	out := append([]byte(nil), data[:12]...)
	for pos := 12; pos < len(data); {
		if len(data)-pos < 8 {
			return nil, fmt.Errorf("truncated WebP chunk")
		}
		n := uint64(binary.LittleEndian.Uint32(data[pos+4 : pos+8]))
		end := uint64(pos) + 8 + n + (n % 2)
		if end > uint64(len(data)) {
			return nil, fmt.Errorf("truncated WebP payload")
		}
		tag := string(data[pos : pos+4])
		if tag != "EXIF" && tag != "XMP " {
			start := len(out)
			out = append(out, data[pos:int(end)]...)
			if tag == "VP8X" && n > 0 {
				out[start+8] &^= 0x0c
			}
		}
		pos = int(end)
	}
	binary.LittleEndian.PutUint32(out[4:8], uint32(len(out)-8))
	return out, nil
}

func stripPNG(data []byte) ([]byte, error) {
	signature := []byte{137, 80, 78, 71, 13, 10, 26, 10}
	if !bytes.HasPrefix(data, signature) {
		return nil, fmt.Errorf("invalid PNG signature")
	}
	out := append([]byte(nil), signature...)
	for pos := 8; pos < len(data); {
		if len(data)-pos < 12 {
			return nil, fmt.Errorf("truncated PNG chunk")
		}
		n := uint64(binary.BigEndian.Uint32(data[pos : pos+4]))
		end := uint64(pos) + 12 + n
		if end > uint64(len(data)) {
			return nil, fmt.Errorf("truncated PNG payload")
		}
		tag := string(data[pos+4 : pos+8])
		switch tag {
		case "eXIf", "tEXt", "zTXt", "iTXt", "tIME":
		default:
			out = append(out, data[pos:int(end)]...)
		}
		pos = int(end)
		if tag == "IEND" {
			return out, nil
		}
	}
	return nil, fmt.Errorf("missing PNG end")
}

func stripJPEG(data []byte) ([]byte, error) {
	if len(data) < 2 || data[0] != 0xff || data[1] != 0xd8 {
		return nil, fmt.Errorf("invalid JPEG signature")
	}
	out := append([]byte(nil), data[:2]...)
	for pos := 2; pos < len(data); {
		start := pos
		if data[pos] != 0xff {
			return nil, fmt.Errorf("invalid JPEG marker")
		}
		for pos < len(data) && data[pos] == 0xff {
			pos++
		}
		if pos >= len(data) {
			return nil, fmt.Errorf("truncated JPEG marker")
		}
		tag := data[pos]
		pos++
		if tag == 0xda || tag == 0xd9 {
			out = append(out, data[start:]...)
			return out, nil
		}
		if tag == 0x01 || (tag >= 0xd0 && tag <= 0xd7) {
			out = append(out, data[start:pos]...)
			continue
		}
		if len(data)-pos < 2 {
			return nil, fmt.Errorf("truncated JPEG length")
		}
		n := int(binary.BigEndian.Uint16(data[pos : pos+2]))
		end := pos + n
		if n < 2 || end > len(data) {
			return nil, fmt.Errorf("truncated JPEG payload")
		}
		// APP1 holds EXIF/XMP; APP13 holds IPTC; COM holds arbitrary text.
		if tag != 0xe1 && tag != 0xed && tag != 0xfe {
			out = append(out, data[start:end]...)
		}
		pos = end
	}
	return nil, fmt.Errorf("missing JPEG image data")
}
