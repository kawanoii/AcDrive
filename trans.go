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
)

//UploadSuccess 上传图片成功时返回的json
type UploadSuccess struct {
	Key string `json:"key"`
}

//UploadFail 上传图片失败时返回的json
type UploadFail struct {
	Error string `json:"error"`
}

func imageUpload(bmp []byte, uptoken string, bmpName string) (string, error) {
	var buff bytes.Buffer
	writer := multipart.NewWriter(&buff)
	w, _ := writer.CreateFormFile("file", bmpName)
	w.Write(bmp)
	writer.WriteField("token", uptoken)
	writer.WriteField("name", bmpName)
	writer.WriteField("key", bmpName) //可自定义，决定最终图片 url
	writer.Close()
	var client http.Client
	req, err := http.NewRequest(
		http.MethodPost,
		"https://up.qbox.me/",
		&buff)

	if err != nil {
		panic(err)
	}
	req.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; Wisn64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.88 Safari/537.36")
	req.Header.Add("Origin", "https://www.acfun.cn")
	req.Header.Add("Referer", "https://www.acfun.cn/member/")
	req.Header.Add("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	bbody, _ := ioutil.ReadAll(resp.Body)
	// fmt.Println(resp.StatusCode)
	if resp.StatusCode != 200 {
		var uploadresp UploadFail
		json.Unmarshal(bbody, &uploadresp)
		return "", errors.New(uploadresp.Error)
	}
	var uploadresp UploadSuccess
	json.Unmarshal(bbody, &uploadresp)
	return uploadresp.Key, nil
}

func upload(filename string, blockSize int, thread int, cookies Cookies) (string, error) {
	skip := func(key string) bool {
		res, err := http.Get(key)
		if err != nil || res.StatusCode != 200 {
			return false
		}
		return true
	}
	core := func(dataBlock dataBlock, cookies Cookies) (BlockMeta, error) {
		var blockMeta BlockMeta
		uptoken, err := getUpToken(cookies)
		if err != nil {
			fmt.Println("上传第", dataBlock.index, "块获取Token时出错", err)
			return blockMeta, err
		}
		bmp := makeBmp(dataBlock.data)
		bmpName := "block_" + dataBlock.sha1
		key, err := imageUpload(bmp, uptoken, bmpName)
		for index := 0; index < 7; index++ {
			if err == nil {
				break
			}
			key, err = imageUpload(bmp, uptoken, bmpName)
			index++
			fmt.Println("第", dataBlock.index, "块上传出错,重试", index, "/ 7", "原因：", err)
		}
		if err != nil {
			return blockMeta, err
		}
		blockMeta = BlockMeta{dataBlock.index, makeURL(key), dataBlock.offset, dataBlock.sha1, dataBlock.size}
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
	if skip(makeURL("meta_" + fileSha1)) {
		fmt.Println(fileInfo.Name(), "秒传成功！")
		// fmt.Println(nCoV(makeURL("meta_" + fileSha1)))
		return "meta_" + fileSha1, nil
	}
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
				if skip(makeURL("block_" + dataBlock.sha1)) {
					mutex.Lock()
					meta.Block[dataBlock.index] = BlockMeta{dataBlock.index, makeURL("block_" + dataBlock.sha1), dataBlock.offset, dataBlock.sha1, dataBlock.size}
					mutex.Unlock()
					fmt.Println("第", dataBlock.index, "块秒传成功！")
					continue
				}
				blockMeta, err := core(dataBlock, cookies)
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
	uptoken, err := getUpToken(cookies)
	if err != nil {
		fmt.Println("上传Meta获取Token时错误。", err)
		return "", err
	}
	metaBmp := makeBmp([]byte(metaJSON))
	bmpName := "meta_" + meta.Sha1
	metakey, err := imageUpload(metaBmp, uptoken, bmpName)
	for index := 0; index < 7; index++ {
		if err == nil {
			break
		}
		metakey, err = imageUpload(metaBmp, uptoken, bmpName)
		index++
		fmt.Println("Meta上传出错,重试", index, "/ 7", "原因:", err)
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
		blockDataSha1 := sha1.Sum(blockData[62:])
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
			blockDataSha1 = sha1.Sum(blockData)
			blockDataSha1Hex = hex.EncodeToString(blockDataSha1[:])

		}
		if err != nil {
			return err
		}
		mutex.Lock()
		f.WriteAt(blockData[62:], blockMeta.Offset)
		mutex.Unlock()
		return nil

	}

	var fmutex sync.RWMutex
	var hmutex sync.RWMutex
	var wg sync.WaitGroup
	blockMetach := make(chan BlockMeta)
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
					fmt.Println("第", blockMeta.Index, "块下载失败，已跳过。")
					continue
				}
				hmutex.Lock()
				history.BlockIndex = append(history.BlockIndex, blockMeta.Index)
				wHistory(history)
				hmutex.Unlock()
				fmt.Println("第", blockMeta.Index, "块下载完成。")
			}
		}()
	}
	for _, blockMeta := range meta.Block {
		blockMetach <- blockMeta
	}
	close(blockMetach)
	wg.Wait()
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
