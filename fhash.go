package main

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"os"
)

type dataBlock struct {
	data   []byte
	index  int
	offset int64
	sha1   string
	size   int
}

func calcSha1(f *os.File) string {
	defer f.Seek(0, 0)
	size := int(4 * 1024 * 1024)

	f.Seek(0, 0)
	sha1 := sha1.New()
	buf := make([]byte, size)
	for n, _ := f.Read(buf); n > 0; n, _ = f.Read(buf) {
		io.WriteString(sha1, string(buf[:n]))
	}
	return hex.EncodeToString(sha1.Sum(nil))
}

func readInChunk(f *os.File, dataBlockch chan<- dataBlock, size int) {
	defer close(dataBlockch)
	defer f.Seek(0, 0)
	f.Seek(0, 0)
	buf := make([]byte, size)
	offset := int64(0)
	index := 0
	for n, _ := f.Read(buf); n > 0; n, _ = f.Read(buf) {
		copybuf := make([]byte, size)
		copy(copybuf, buf)
		sha1 := sha1.Sum(copybuf)
		dataBlock := dataBlock{copybuf, index, offset, hex.EncodeToString(sha1[:]), n}
		dataBlockch <- dataBlock
		offset += int64(n)
		index++
	}
}
