package main

//Status 运行状态
type Status struct {
	AcAuth int          `json:"acauth"`
	Downs  []DownStatus `json:"downs"`
	Ups    []UpStatus   `json:"ups"`
}

//Info 解析信息
type Info struct {
	Filename   string `json:"filename"`
	FileSha1   string `json:"filesha1"`
	FileSize   string `json:"size"`
	BlockNum   int    `json:"blocknum"`
	UploadTime string `json:"uploadtime"`
}

//DownStatus 下载状态
type DownStatus struct {
	Filename string `json:"filename"`
	FileSha1 string `json:"filesha1"`
	FileSize int64  `json:"size"`
	BlockNum int    `json:"blocknum"`
	OKNUM    int    `json:"oknum"`
	Code     int    `json:"code"`
	Error    error  `json:"error"`
	Message  string `json:"message"`
}

//UpStatus 上传状态
type UpStatus struct {
	Filename string `json:"filename"`
	FileSha1 string `json:"filesha1"`
	FileSize int64  `json:"size"`
	BlockNum int    `json:"blocknum"`
	OKNUM    int    `json:"oknum"`
	Code     int    `json:"code"`
	Error    error  `json:"error"`
	Ncd      string `json:"ncd"`
	Message  string `json:"message"`
}

//UpReq 上传请求
type UpReq struct {
	FilePath  string `json:"filepath"`
	BlockSize int    `json:"blocksize"`
	Thread    int    `json:"thread"`
}

//DownReq 下载请求
type DownReq struct {
	DownPath string `json:"downpath"`
	Ncd      string `json:"ncd"`
	Thread   int    `json:"thread"`
}

//AcUser Acfun用户密码
type AcUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
