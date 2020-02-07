package main

import (
	"errors"
	"strconv"
)

// alidrive://ae01.alicdn.com/kf/H64e6310518a54d4b8936e8f5105e0ed9T
// https://ae01.alicdn.com/kf/H64e6310518a54d4b8936e8f5105e0ed9T.bmp

func makeMetaURL(aliurl string) string {
	return "alidrive://" + aliurl[27:][:34]
}

func unMetaURL(metaurl string) (string, error) {
	if len(metaurl) != 45 {
		return "", errors.New("Meta URL 长度不符。")
	}

	if metaurl[:11] != "alidrive://" {
		return "", errors.New("Meta URL 通常以 alidrive:// 开头。")

	}
	return makeURL(metaurl[11:]), nil
}

func makeURL(key string) string {
	return "https://ae01.alicdn.com/kf/" + key + ".bmp"
}

func sizeString(byteint int64) string {
	byte := float64(byteint)
	if byte > 1024*1024*1024 {
		return strconv.FormatFloat(byte/1024/1024/1024, 'f', 3, 64) + " GB"
	} else if byte > 1024*1024 {
		return strconv.FormatFloat(byte/1024/1024, 'f', 3, 64) + " MB"

	} else if byte > 1024 {
		return strconv.FormatFloat(byte/1024, 'f', 3, 64) + " KB"

	} else {
		return strconv.FormatFloat(byte, 'f', 3, 64) + " B"

	}
}
