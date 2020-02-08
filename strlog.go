package main

import (
	"errors"
	"strconv"
)

func makeMetaURL(key string) string {
	return "acdrive://" + key
}

func unMetaURL(metaurl string) (string, error) {
	if len(metaurl) != 55 {
		return "", errors.New("Meta URL 长度不符。")
	}

	if metaurl[:10] != "acdrive://" {
		return "", errors.New("Meta URL 通常以 acdrive:// 开头。")

	}
	return makeURL(metaurl[10:]), nil
}

func makeURL(key string) string {
	return "https://imgs.aixifan.com/" + key
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
