package ui

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"log"
	"unsafe"

	"golang.org/x/sys/windows"
)

//go:embed assets/icon.png
var iconPNG []byte

var iconData []byte

func init() {
	ico, err := pngToICO(iconPNG)
	if err != nil {
		log.Printf("icon: failed to build ICO from PNG, using fallback: %v", err)
		iconData = buildFallbackICO(32)
		return
	}
	iconData = ico
}

func pngToICO(pngData []byte) ([]byte, error) {
	src, _, err := image.Decode(bytes.NewReader(pngData))
	if err != nil {
		return nil, err
	}

	sizes := []int{16, 32, 48, 256}
	var chunks [][]byte
	for _, sz := range sizes {
		resized := resizeImage(src, sz)
		var buf bytes.Buffer
		if err := png.Encode(&buf, resized); err != nil {
			return nil, err
		}
		chunks = append(chunks, buf.Bytes())
	}
	return buildICOFromPNGs(sizes, chunks), nil
}

func resizeImage(src image.Image, size int) *image.NRGBA {
	dst := image.NewNRGBA(image.Rect(0, 0, size, size))
	b := src.Bounds()
	srcW := b.Max.X - b.Min.X
	srcH := b.Max.Y - b.Min.Y
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			sx := b.Min.X + x*srcW/size
			sy := b.Min.Y + y*srcH/size
			dst.Set(x, y, src.At(sx, sy))
		}
	}
	return dst
}

func buildICOFromPNGs(sizes []int, chunks [][]byte) []byte {
	n := len(sizes)
	headerSize := 6 + n*16

	offsets := make([]int, n)
	total := headerSize
	for i, c := range chunks {
		offsets[i] = total
		total += len(c)
	}

	buf := make([]byte, total)

	binary.LittleEndian.PutUint16(buf[0:], 0) // reserved
	binary.LittleEndian.PutUint16(buf[2:], 1) // type = 1 (ICO)
	binary.LittleEndian.PutUint16(buf[4:], uint16(n))

	// Directory entries
	for i, sz := range sizes {
		o := 6 + i*16
		w, h := byte(sz), byte(sz)
		if sz >= 256 {
			w, h = 0, 0 // 0 encodes 256
		}
		buf[o+0] = w
		buf[o+1] = h
		buf[o+2] = 0                                 // colour count (0 = no palette)
		buf[o+3] = 0                                 // reserved
		binary.LittleEndian.PutUint16(buf[o+4:], 1)  // colour planes
		binary.LittleEndian.PutUint16(buf[o+6:], 32) // bits per pixel
		binary.LittleEndian.PutUint32(buf[o+8:], uint32(len(chunks[i])))
		binary.LittleEndian.PutUint32(buf[o+12:], uint32(offsets[i]))
		copy(buf[offsets[i]:], chunks[i])
	}
	return buf
}

var (
	modUser32Icon           = windows.NewLazySystemDLL("user32.dll")
	procCreateIconFromResEx = modUser32Icon.NewProc("CreateIconFromResourceEx")
	procSendMessageIcon     = modUser32Icon.NewProc("SendMessageW")
)

// SetWindowIcon sets the HICON for the title bar and taskbar button of hwnd.
// It is safe to call from any goroutine as long as the window has been created.
func SetWindowIcon(hwnd windows.HWND) {
	if len(iconPNG) == 0 {
		return
	}
	src, _, err := image.Decode(bytes.NewReader(iconPNG))
	if err != nil {
		log.Printf("icon: SetWindowIcon: decode: %v", err)
		return
	}
	for _, pair := range []struct{ sz, kind uintptr }{{16, 0}, {32, 1}} {
		var buf bytes.Buffer
		if err := png.Encode(&buf, resizeImage(src, int(pair.sz))); err != nil {
			continue
		}
		b := buf.Bytes()
		// CreateIconFromResourceEx accepts raw PNG bytes on Vista+.
		h, _, _ := procCreateIconFromResEx.Call(
			uintptr(unsafe.Pointer(&b[0])),
			uintptr(len(b)),
			1,          // fIcon = TRUE
			0x00030000, // dwVersion = 3.0
			pair.sz, pair.sz,
			0, // LR_DEFAULTCOLOR
		)
		if h != 0 {
			const wmSetIcon = 0x0080
			procSendMessageIcon.Call(uintptr(hwnd), wmSetIcon, pair.kind, h)
		}
	}
}

// buildFallbackICO returns a plain solid-colour ICO used only when the
// embedded PNG cannot be decoded.
func buildFallbackICO(size int) []byte {
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	fill := color.NRGBA{R: 0x5B, G: 0x7C, B: 0x99, A: 0xFF}
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img.SetNRGBA(x, y, fill)
		}
	}
	var buf bytes.Buffer
	png.Encode(&buf, img) //nolint:errcheck
	return buildICOFromPNGs([]int{size}, [][]byte{buf.Bytes()})
}
