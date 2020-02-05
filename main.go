package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Println("请使用 'login' , 'download' , 'upload' 或 'info' 子命令。\"-h\" 参数查看帮助。\n例如 acdrive login -h \n查看登录的帮助。")
		os.Exit(1)
	}

	loginCmd := flag.NewFlagSet("login", flag.ExitOnError)
	username := loginCmd.String("u", "", "用户名")
	password := loginCmd.String("p", "", "密码")

	uploadCmd := flag.NewFlagSet("upload", flag.ExitOnError)
	filename := uploadCmd.String("f", "", "上传文件名")
	upthread := uploadCmd.Int("t", 4, "上传线程数")
	blocksize := uploadCmd.Int("bs", 4, "文件分块大小")

	downloadCmd := flag.NewFlagSet("download", flag.ExitOnError)
	downmetaurl := downloadCmd.String("m", "", "Meta URL,通常以\"acdrive://\"开头")
	downthread := downloadCmd.Int("t", 4, "下载线程数")

	infoCmd := flag.NewFlagSet("info", flag.ExitOnError)
	infometaurl := infoCmd.String("m", "", "Meta URL,通常以\"acdrive://\"开头")

	switch os.Args[1] {
	case "login":
		loginCmd.Parse(os.Args[2:])
		ck, err := login(*username, *password)
		if err != nil {
			fmt.Println(err)
			return
		}
		wCookie(ck)
	case "upload":
		uploadCmd.Parse(os.Args[2:])
		ck, err := rCookie()
		if err != nil {
			fmt.Println(err)
			return
		}
		metakey, err := upload(*filename, *blocksize*1024*1024, *upthread, ck)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("地址：")
		fmt.Println(makeMetaURL(metakey))
	case "download":
		downloadCmd.Parse(os.Args[2:])
		download(*downmetaurl, *downthread)
	case "info":
		infoCmd.Parse(os.Args[2:])
		infoMeta(*infometaurl)
	default:
		fmt.Println("请使用 'login' , 'download' , 'upload' 或 'info' 子命令。\"-h\" 参数查看帮助。\n例如 acdrive login -h \n查看登录的帮助。")
		os.Exit(1)

	}
}
