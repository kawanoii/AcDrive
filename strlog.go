package main

import (
	"encoding/base64"
	"errors"
	"strconv"
)

func nCoV(url string) string {
	return "nCoVDrive-" + base64.StdEncoding.EncodeToString([]byte(url))
}

func unnCov(ncd string) (string, error) {
	if len(ncd) < 11 {
		return "", errors.New("nCoVDrive 链接过短")
	}
	decodeBytes, err := base64.StdEncoding.DecodeString(ncd[10:])
	return string(decodeBytes), err
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

func makeURL(key string) string {
	return "https://imgs.aixifan.com/" + key
}
