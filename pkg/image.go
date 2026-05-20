package pkg

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/jpeg"
	"image/png"
	"io"

	"golang.org/x/image/draw"
)

func ResizeImage(src image.Image, sz int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, sz, sz))
	draw.BiLinear.Scale(dst, dst.Rect.Bounds(), src, src.Bounds(), draw.Over, nil)
	return dst
}

func DecodeImage(r io.ReadSeeker) (image.Image, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		r.Seek(0, io.SeekStart)
		img, err = jpeg.Decode(r)
	}

	return img, err
}

func EncodeICO(w io.Writer, img image.Image) error {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return err
	}

	png := buf.Bytes()

	b := img.Bounds()
	sz := b.Dx()
	wd := uint8(sz)
	if sz == 256 {
		wd = 0
	}

	binary.Write(w, binary.LittleEndian, uint16(0))
	binary.Write(w, binary.LittleEndian, uint16(1))
	binary.Write(w, binary.LittleEndian, uint16(1))

	binary.Write(w, binary.LittleEndian, wd)
	binary.Write(w, binary.LittleEndian, wd)
	binary.Write(w, binary.LittleEndian, uint8(0))
	binary.Write(w, binary.LittleEndian, uint8(0))
	binary.Write(w, binary.LittleEndian, uint16(1))
	binary.Write(w, binary.LittleEndian, uint16(32))
	binary.Write(w, binary.LittleEndian, uint32(len(png)))
	binary.Write(w, binary.LittleEndian, uint32(6+16))

	w.Write(png)
	return nil
}
