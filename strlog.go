package main

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"time"
)

func nCoV(url string) string {
	return "nCoVDrive://" + base64.StdEncoding.EncodeToString([]byte(url))
}

func unnCov(ncd string) (string, error) {
	decodeBytes, err := base64.StdEncoding.DecodeString(ncd[12:])
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

func log(message string) {
	fmt.Println(time.Now(), message)
}
