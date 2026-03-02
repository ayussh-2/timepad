package ui

import "encoding/binary"

var iconData = buildICO(32, 0x99, 0x7C, 0x5B)

func buildICO(size int, b, g, r byte) []byte {
	pixels := make([]byte, size*size*4)
	for i := 0; i < size*size; i++ {
		pixels[i*4+0] = b
		pixels[i*4+1] = g
		pixels[i*4+2] = r
		pixels[i*4+3] = 0xFF
	}

	maskRowBytes := ((size + 31) / 32) * 4
	mask := make([]byte, size*maskRowBytes)

	bmpSize := 40 + len(pixels) + len(mask)
	buf := make([]byte, 22+bmpSize)

	binary.LittleEndian.PutUint16(buf[0:], 0)
	binary.LittleEndian.PutUint16(buf[2:], 1)
	binary.LittleEndian.PutUint16(buf[4:], 1)

	buf[6] = byte(size)
	buf[7] = byte(size)
	binary.LittleEndian.PutUint16(buf[10:], 1)
	binary.LittleEndian.PutUint16(buf[12:], 32)
	binary.LittleEndian.PutUint32(buf[14:], uint32(bmpSize))
	binary.LittleEndian.PutUint32(buf[18:], 22)

	o := 22
	binary.LittleEndian.PutUint32(buf[o:], 40)
	binary.LittleEndian.PutUint32(buf[o+4:], uint32(size))
	binary.LittleEndian.PutUint32(buf[o+8:], uint32(size*2))
	binary.LittleEndian.PutUint16(buf[o+12:], 1)
	binary.LittleEndian.PutUint16(buf[o+14:], 32)
	binary.LittleEndian.PutUint32(buf[o+20:], uint32(len(pixels)+len(mask)))
	o += 40
	copy(buf[o:], pixels)
	o += len(pixels)
	copy(buf[o:], mask)
	return buf
}
