package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/Si-Huan/mkbmp"
)

//UploadR 上传图片成功时返回的json
type UploadR struct {
	Code   int    `json:"code"`
	ImgURL string `json:"imgurl"`
	Msg    string `json:"msg"`
}

func imageUpload(bmp []byte, sha1 string) (string, error) {
	var buff bytes.Buffer
	writer := multipart.NewWriter(&buff)
	w, _ := writer.CreateFormFile("Filedata", sha1+".bmp")
	w.Write(bmp)
	writer.WriteField("file", "multipart")
	writer.Close()
	var client http.Client
	req, err := http.NewRequest(
		http.MethodPost,
		"https://api.uomg.com/api/image.ali",
		&buff)

	if err != nil {
		panic(err)
	}
	req.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; Wisn64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.88 Safari/537.36")
	req.Header.Add("Origin", "https://www.taobao.com")
	req.Header.Add("Referer", "https://www.taobao.com/")
	req.Header.Add("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	bbody, _ := ioutil.ReadAll(resp.Body)
	// fmt.Println(string(bbody))
	if resp.StatusCode != 200 {
		return "", errors.New("非200返回")
	}
	var uploadresp UploadR
	json.Unmarshal(bbody, &uploadresp)
	if uploadresp.Code != 1 {
		// fmt.Println(uploadresp.Msg)
		return "", errors.New("上传失败,msg:" + uploadresp.Msg)
	}
	if uploadresp.ImgURL == "" {
		return "", errors.New("图片地址为NULL")
	}
	return uploadresp.ImgURL, nil
}

func upload(filename string, blockSize int, thread int) (string, error) {
	core := func(dataBlock dataBlock) (BlockMeta, error) {
		var blockMeta BlockMeta
		bmp := mkalibmp(dataBlock.data)
		imgurl, err := imageUpload(bmp, dataBlock.sha1)
		for index := 0; index < 1000; index++ {
			if err == nil {
				break
			}
			imgurl, err = imageUpload(bmp, dataBlock.sha1)
			index++
			fmt.Println("第", dataBlock.index, "块上传出错,重试", index, "/ 1000", "原因：", err)
		}
		if err != nil {
			return blockMeta, err
		}
		blockMeta = BlockMeta{dataBlock.index, imgurl, dataBlock.offset, dataBlock.sha1, dataBlock.size}
		return blockMeta, nil
	}

	f, err := os.Open(filename)
	if err != nil {
		fmt.Println("打开文件时出错。", err)
		return "", err
	}
	defer f.Close()
	fileInfo, _ := f.Stat()
	fmt.Println("开始上传")
	fmt.Println("计算校验和。")
	fileSha1 := calcSha1(f)
	fmt.Println("计算完毕。")
	var mutex sync.RWMutex
	var wg sync.WaitGroup
	flag := false
	allBlock := math.Ceil(float64(fileInfo.Size()) / float64(blockSize))
	fmt.Println("共", allBlock, "块")
	blocks := make([]BlockMeta, int(allBlock))
	meta := Meta{Block: blocks}
	dataBlockch := make(chan dataBlock)

	for i := 0; i < thread; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for dataBlock := range dataBlockch {
				blockMeta, err := core(dataBlock)
				if err != nil {
					fmt.Println("第", dataBlock.index, "块上传失败，已跳过。")
					flag = true
					continue
				}
				fmt.Println("第", blockMeta.Index, "块上传完成。")
				mutex.Lock()
				meta.Block[dataBlock.index] = blockMeta
				mutex.Unlock()
			}
		}()
	}
	readInChunk(f, dataBlockch, blockSize)
	wg.Wait()
	if flag {
		return "", errors.New("有块未上传成功，故未上传Meta，请重试！")
	}
	meta.Time = time.Now().Unix()
	meta.Filename = fileInfo.Name()
	meta.Size = fileInfo.Size()
	meta.Sha1 = fileSha1
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	metaBmp := mkbmp.MakeBmp([]byte(metaJSON))
	metakey, err := imageUpload(metaBmp, meta.Sha1)
	for index := 0; index < 100; index++ {
		if err == nil {
			break
		}
		metakey, err = imageUpload(metaBmp, meta.Sha1)
		index++
		fmt.Println("Meta上传出错,重试", index, "/ 100", "原因:", err)
	}
	if err != nil {
		return "", err
	}
	fmt.Println("上传完毕！")
	return metakey, nil
}

func imageDownload(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return []byte(""), err
	}
	if res.StatusCode != 200 {
		return []byte(""), errors.New("下载错误！StatusCode:" + string(res.StatusCode))
	}
	imagedata, _ := ioutil.ReadAll(res.Body)
	return imagedata, nil
}

func download(ncd string, thread int) {
	skip := func(index int, historyIndex []int) bool {
		for _, hIndex := range historyIndex {
			if index == hIndex {
				return true
			}
		}
		return false
	}
	core := func(blockMeta BlockMeta, mutex *sync.RWMutex, f *os.File) error {
		blockData, err := imageDownload(blockMeta.URL)
		blockDataSha1 := sha1.Sum(blockData[54:])
		blockDataSha1Hex := hex.EncodeToString(blockDataSha1[:])
		for index := 0; index < 7; index++ {
			if err == nil && blockDataSha1Hex == blockMeta.Sha1 {
				break
			}
			if err == nil {
				err = errors.New("Sha1校验失败。")
			}
			fmt.Println("第", blockMeta.Index, "块下载出错,重试", index, "/ 7", "原因：", err)
			blockData, err = imageDownload(blockMeta.URL)
			blockDataSha1 = sha1.Sum(blockData[54:])
			blockDataSha1Hex = hex.EncodeToString(blockDataSha1[:])

		}
		if err != nil {
			return err
		}
		mutex.Lock()
		f.WriteAt(blockData[54:][:blockMeta.Size], blockMeta.Offset)
		mutex.Unlock()
		return nil

	}

	var fmutex sync.RWMutex
	var hmutex sync.RWMutex
	var wg sync.WaitGroup
	blockMetach := make(chan BlockMeta)
	fmt.Println("解析Meta")
	meta, err := getMeta(ncd)
	if err != nil {
		fmt.Println("获取Meta出错。", err)
		return
	}
	history, err := rHistory(meta.Sha1)
	if err != nil {
		history.FileSha1 = meta.Sha1
	}
	f, err := os.Open(meta.Filename)
	if err != nil {
		f, err = os.Create(meta.Filename)
	}
	if err != nil {
		fmt.Println("创建文件时出错。", err)
		return
	}
	// fmt.Println(meta)
	fmt.Println("开始下载！")
	for i := 0; i < thread; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for blockMeta := range blockMetach {
				if skip(blockMeta.Index, history.BlockIndex) {
					fmt.Println("第", blockMeta.Index, "块已下载，跳过")
					continue
				}
				err := core(blockMeta, &fmutex, f)
				if err != nil {
					fmt.Println("第", blockMeta.Index, "块下载失败，已跳过。错误信息:", err.Error())
					continue
				}
				hmutex.Lock()
				history.BlockIndex = append(history.BlockIndex, blockMeta.Index)
				wHistory(history)
				hmutex.Unlock()
				fmt.Println("第", blockMeta.Index, "块下载完成并通过校验。")
			}
		}()
	}
	for _, blockMeta := range meta.Block {
		blockMetach <- blockMeta
	}
	close(blockMetach)
	wg.Wait()
	fmt.Println("文件下载全部完成，在在对整个文件进行校验.\n若文件较大,可能会需要一段时间.\n如果你想跳过校验的话，直接结束程序就好啦.")
	fileSha1 := calcSha1(f)
	if fileSha1 != meta.Sha1 {
		fmt.Println("文件校验失败，下载失败！请重试！")
		return
	}
	fmt.Println("文件校验通过，下载完成！")
}

func infoMeta(ncd string) {
	meta, err := getMeta(ncd)
	if err != nil {
		fmt.Println("获取Meta出错。", err)
		return
	}
	tm := time.Unix(meta.Time, 0)
	fmt.Println("文件名称： ", meta.Filename)
	fmt.Println("上传时间： ", tm.Format("2006-01-02 15:04:05"))
	fmt.Println("文件大小： ", sizeString(meta.Size))
	fmt.Println("文件哈希：  Sha1", meta.Sha1)
	fmt.Println("文件分块： ", len(meta.Block))
}
