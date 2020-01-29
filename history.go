package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

//History 本地下载历史 用于续传
type History struct {
	FileSha1   string `json:"sha1"`
	BlockIndex []int  `json:"block"`
}

func wHistory(history History) {
	historyJSON, err := json.Marshal(history)
	if err != nil {
		fmt.Println(err)
		return
	}
	f, err := os.Create(history.FileSha1 + ".json")
	if err != nil {
		fmt.Println("创建记录下载历史文件时出错。", err)
		return
	}
	_, err = f.Write(historyJSON)
	if err != nil {
		fmt.Println("写入下载历史出错。", err)
		return
	}
}

func rHistory(filesha1 string) (History, error) {
	var history History
	f, err := os.Open(filesha1 + ".json")
	if err != nil {
		return history, errors.New("打开历史出错。没有历史文件。")
	}
	bhistory, err := ioutil.ReadAll(f)
	if err != nil {
		return history, errors.New("读取history出错。")
	}
	err = json.Unmarshal(bhistory, &history)
	return history, err
}
