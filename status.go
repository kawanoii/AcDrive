package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

//Status 运行状态
type Status struct {
	AcAuth int                   `json:"acauth"`
	Downs  map[string]DownStatus `json:"downs"`
	Ups    map[string]UpStatus   `json:"ups"`
}

//Info 解析信息
type Info struct {
	Filename   string `json:"filename"`
	FileSha1   string `json:"filesha1"`
	FileSize   string `json:"size"`
	BlockNum   int    `json:"blocknum"`
	UploadTime string `json:"uploadtime"`
	Error      string `json:"error"`
}

//DownStatus 下载状态
type DownStatus struct {
	Ncd      string `json:"ncd"`
	Filename string `json:"filename"`
	FileSha1 string `json:"filesha1"`
	FileSize string `json:"size"`
	BlockNum int    `json:"blocknum"`
	OKNUM    int32  `json:"oknum"`
	Code     int    `json:"code"`
	Error    string `json:"error"`
	Message  string `json:"message"`
}

//UpStatus 上传状态
type UpStatus struct {
	Filename string `json:"filename"`
	FileSha1 string `json:"filesha1"`
	FileSize int64  `json:"size"`
	BlockNum int    `json:"blocknum"`
	OKNUM    int32  `json:"oknum"`
	Code     int    `json:"code"`
	Error    error  `json:"error"`
	Ncd      string `json:"ncd"`
	Message  string `json:"message"`
}

//UpReq 上传请求
type UpReq struct {
	FileName  string `json:"filename"`
	BlockSize int    `json:"blocksize"`
	Thread    int    `json:"thread"`
}

//DownReq 下载请求
type DownReq struct {
	Ncd    string `json:"ncd"`
	Thread int    `json:"thread"`
}

//AcUser Acfun用户密码
type AcUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func wStatus(status Status) {
	statusJSON, err := json.Marshal(status)
	if err != nil {
		fmt.Println(err)
		return
	}
	f, err := os.Create("status.json")
	if err != nil {
		fmt.Println("创建Status文件时出错。", err)
		return
	}
	_, err = f.Write(statusJSON)
	if err != nil {
		fmt.Println("写入Status文件出错。", err)
		return
	}
}

func rStatus() Status {
	var status Status
	status.Downs = make(map[string]DownStatus)
	status.Ups = make(map[string]UpStatus)
	f, err := os.Open("status.json")
	if err != nil {
		return status
	}
	bstatus, err := ioutil.ReadAll(f)
	if err != nil {
		return status
	}
	json.Unmarshal(bstatus, &status)
	return status
}
