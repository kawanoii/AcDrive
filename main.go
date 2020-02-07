package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {

	// fmt.Println(upload("ClashX.dmg", 1228800, 4))
	// download("alidrive://H64e6310518a54d4b8936e8f5105e0ed9T", 1)
	if len(os.Args) < 2 {
		fmt.Println("请使用 'download' , 'upload' 或 'info' 子命令。\"-h\" 参数查看帮助。\n例如 alidrive download -h \n查看下载的帮助。")
		os.Exit(1)
	}

	uploadCmd := flag.NewFlagSet("upload", flag.ExitOnError)
	filename := uploadCmd.String("f", "", "上传文件名")
	upthread := uploadCmd.Int("t", 4, "上传线程数")
	// blocksize := uploadCmd.Int("bs", 4, "文件分块大小")

	downloadCmd := flag.NewFlagSet("download", flag.ExitOnError)
	downmetaurl := downloadCmd.String("m", "", "Meta URL,通常以\"alidrive://\"开头")
	downthread := downloadCmd.Int("t", 4, "下载线程数")

	infoCmd := flag.NewFlagSet("info", flag.ExitOnError)
	infometaurl := infoCmd.String("m", "", "Meta URL,通常以\"alidrive://\"开头")

	switch os.Args[1] {
	case "upload":
		uploadCmd.Parse(os.Args[2:])
		metakey, err := upload(*filename, 1228800, *upthread)
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
		fmt.Println("download' , 'upload' 或 'info' 子命令。\"-h\" 参数查看帮助。\n例如 alidrive download -h \n查看下载的帮助。")
		os.Exit(1)

	}
}
