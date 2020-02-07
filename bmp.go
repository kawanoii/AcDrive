package main

//Bmp 没有考虑调色板 24 和 36的不需要调色板
type Bmp struct {
	Header BitmapFileHeader
	DIB    DIBHeader
	Data   []byte
}

//BitmapFileHeader bmp图头
type BitmapFileHeader struct {
	BM                     [2]byte //头文件字段,通常为"BM"
	FullSize               [4]byte //整个bmp图片大小
	SomethingNotImportant1 [2]byte //预留字段
	SomethingNotImportant2 [2]byte //预留字段
	Start                  [4]byte //图片信息开始的地方
}

// DIBHeader To store detailed information about the bitmap image and define the pixel format
type DIBHeader struct {
	DIBSize                [4]byte //DIB 大小通常为 0x28
	Width                  [4]byte
	Height                 [4]byte
	ColorPlane             [2]byte //色彩平面数量,必须为 1
	PerPixelBit            [2]byte //每像素占的位数
	CompressionMethod      [4]byte //压缩方式,通常不压缩,对应 0
	DataSize               [4]byte //原始位图大小,对于不压缩的设置为 0
	SomethingNotImportant1 [4]byte //横向分辨率,像素每米
	SomethingNotImportant2 [4]byte //纵向分辨率,像素每米
	SomethingNotImportant3 [4]byte //调色板颜色数量,通常为 0. 不代表没有颜色
	SomethingNotImportant4 [4]byte //重要颜色数量,通常被忽略,为 0 时代表所有颜色都重要
}

func (bfh *BitmapFileHeader) b() []byte {
	var b []byte
	b = append([]byte(""), bfh.BM[:]...)
	return b
}

func mkalibmp(data []byte) []byte {
	header := []byte{66, 77, 54, 192, 18, 0, 0, 0, 0, 0, 54, 0, 0, 0, 40, 0, 0, 0, 128, 2, 0, 0, 224, 1, 0, 0, 1, 0, 32, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	return append(header, data...)
}
