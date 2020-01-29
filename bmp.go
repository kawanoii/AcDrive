package main

import (
	"math"
)

func makeBmp(data []byte) []byte {
	header := []byte{66, 77, 0, 0, 0, 0, 0, 0, 0, 0, 62, 0, 0, 0, 40, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0, 1, 0, 0, 0, 0, 0, 175, 30, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 255, 255, 255, 0}
	bmpsize := uint32(62 + len(data))
	header[2] = uint8(bmpsize)
	header[3] = uint8(bmpsize >> 8)
	header[4] = uint8(bmpsize >> 16)
	header[5] = uint8(bmpsize >> 24)
	bmpwigth := uint32(len(data))
	header[18] = uint8(bmpwigth)
	header[19] = uint8(bmpwigth >> 8)
	header[20] = uint8(bmpwigth >> 16)
	header[21] = uint8(bmpwigth >> 24)
	datasize := uint32(math.Ceil(float64(len(data)) / 8))
	header[34] = uint8(datasize)
	header[35] = uint8(datasize >> 8)
	header[36] = uint8(datasize >> 16)
	header[37] = uint8(datasize >> 24)
	return append(header, data...)
}
