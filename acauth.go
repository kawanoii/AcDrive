package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	// "fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	_ "strings"
)

//Cookies 登陆A站成功的5个cookie
type Cookies [5]http.Cookie

//UpTokenresp 获取uptoken返回的json
type UpTokenresp struct {
	ErrorID   int    `json:"errorid"`
	RequestID string `json:"requestid"`
	ErrorDesc string `json:"errordesc"`
	Vdata     Vdata  `json:"vdata"`
}

//Vdata uptoken 返回字典中的vdata
type Vdata struct {
	UpToken string `json:"uptoken"`
}

// Loginresp 登陆A站的返回json
type Loginresp struct {
	Result int    `json:"result"`
	Error  string `json:"error_msg"`
}

func login(username, password string) (Cookies, error) {
	form := make(url.Values)
	form.Add("username", username)
	form.Add("password", password)
	resp, err := http.PostForm("https://id.app.acfun.cn/rest/web/login/signin", form)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	statuscode := resp.StatusCode
	bbody, _ := ioutil.ReadAll(resp.Body)
	var loginresp Loginresp
	json.Unmarshal(bbody, &loginresp)
	var cookies Cookies
	if statuscode != 200 {
		return cookies, errors.New("网络错误! StatusCode:" + string(statuscode))
	} else if loginresp.Result != 0 {
		return cookies, errors.New("登录错误! ErrorMessage:" + loginresp.Error)
	}
	for index, cookie := range resp.Cookies() {
		cookies[index] = *cookie
	}
	return cookies, nil

}

func getUpToken(cookies Cookies) (string, error) {
	req, err := http.NewRequest(http.MethodGet, "https://www.acfun.cn/v2/user/content/upToken", nil)
	if err != nil {
		panic(err)
	}
	for _, cookie := range cookies {
		req.AddCookie(&cookie)
	}
	req.Header.Add("devicetype", "7")
	req.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.88 Safari/537.36")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	statuscode := resp.StatusCode

	bbody, _ := ioutil.ReadAll(resp.Body)
	var uptokenresp UpTokenresp
	json.Unmarshal(bbody, &uptokenresp)

	if statuscode != 200 {
		return "", errors.New("网络错误! StatusCode:" + string(statuscode))
	} else if uptokenresp.ErrorID != 0 {
		return "", errors.New("获取Token错误! ErrorMessage:" + uptokenresp.ErrorDesc)
	}
	token, _ := base64.StdEncoding.DecodeString(uptokenresp.Vdata.UpToken)
	return string(token)[5:], nil
}

func wCookie(cookies Cookies) {
	ckJSON, err := json.Marshal(cookies)
	if err != nil {
		fmt.Println(err)
		return
	}
	f, err := os.Create("cookies.json")
	if err != nil {
		fmt.Println("创建Cookie文件时出错。", err)
		return
	}
	_, err = f.Write(ckJSON)
	if err != nil {
		fmt.Println("写入Cookie文件出错。", err)
		return
	}
}

func rCookie() (Cookies, error) {
	var cookies Cookies
	f, err := os.Open("cookies.json")
	if err != nil {
		return cookies, errors.New("打开Cookie文件出错。尝试登录。")
	}
	bcookies, err := ioutil.ReadAll(f)
	if err != nil {
		return cookies, errors.New("读取Cookie出错。尝试重新登录。")
	}
	err = json.Unmarshal(bcookies, &cookies)
	return cookies, err
}
