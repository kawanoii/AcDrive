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
	"strconv"
	"sync"
	"sync/atomic"
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

type update struct {
	key   string
	value string
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

func upload(filename string, blockSize int, thread int, cookies Cookies, ups UpStatus, upch chan UpStatus) {
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
			ups.Message = "第" + strconv.Itoa(dataBlock.index) + "块上传出错,重试" + strconv.Itoa(index) + "/ 7 原因：" + err.Error()
			upch <- ups
			fmt.Println("第", dataBlock.index, "块上传出错,重试", index, "/ 7", "原因：", err)
		}
		if err != nil {
			return blockMeta, err
		}
		blockMeta = BlockMeta{dataBlock.index, makeURL(key), dataBlock.offset, dataBlock.sha1, dataBlock.size}
		return blockMeta, nil
	}
	ups.Code = 1
	ups.OKNUM = 0
	upch <- ups
	f, err := os.Open(filename)
	if err != nil {
		ups.Code = -1
		ups.Error = errors.New("打开文件出错" + err.Error())
		fmt.Println("打开文件时出错。", err)
		upch <- ups
		return
	}
	defer f.Close()
	fileInfo, _ := f.Stat()
	ups.Filename = fileInfo.Name()
	ups.FileSize = fileInfo.Size()
	allBlock := int(math.Ceil(float64(fileInfo.Size()) / float64(blockSize)))
	ups.BlockNum = allBlock
	fmt.Println("开始上传")
	ups.Message = "正在计算校验和。"
	upch <- ups
	fmt.Println("计算校验和。")
	fileSha1 := calcSha1(f)
	ups.FileSha1 = fileSha1
	upch <- ups
	fmt.Println("计算完毕。")
	if skip(makeURL("meta_" + fileSha1)) {
		ups.OKNUM = int32(allBlock)
		ups.Code = 0
		ups.Ncd = makeMetaURL("meta_" + fileSha1)
		ups.Message = "秒传成功！"
		fmt.Println(fileInfo.Name(), "秒传成功！")
		upch <- ups
		return
	}
	var mutex sync.RWMutex
	var wg sync.WaitGroup
	flag := false
	fmt.Println("共", allBlock, "块")
	blocks := make([]BlockMeta, allBlock)
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
					atomic.AddInt32(&ups.OKNUM, 1)
					ups.Message = "第" + strconv.Itoa(dataBlock.index) + "块秒传成功！"
					upch <- ups
					fmt.Println("第", dataBlock.index, "块秒传成功！")
					continue
				}
				blockMeta, err := core(dataBlock, cookies)
				if err != nil {
					ups.Message = "第" + strconv.Itoa(dataBlock.index) + "块上传失败，已跳过。"
					upch <- ups
					fmt.Println("第", dataBlock.index, "块上传失败，已跳过。")
					flag = true
					continue
				}
				ups.Message = "第" + strconv.Itoa(blockMeta.Index) + "块上传完成。"
				atomic.AddInt32(&ups.OKNUM, 1)
				upch <- ups
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
		ups.Error = errors.New("有块未上传成功，故未上传Meta，请重新登录再试！")
		ups.Code = -1
		upch <- ups
		return
	}
	meta.Time = time.Now().Unix()
	meta.Filename = fileInfo.Name()
	meta.Size = fileInfo.Size()
	meta.Sha1 = fileSha1
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		ups.Message = err.Error()
		ups.Error = err
		ups.Code = -1
		fmt.Println(err)
		upch <- ups
		return
	}
	uptoken, err := getUpToken(cookies)
	if err != nil {
		ups.Message = err.Error()
		ups.Error = errors.New("上传Meta获取Token时错误。" + err.Error())
		fmt.Println("上传Meta获取Token时错误。", err)
		ups.Code = -1
		upch <- ups
		return
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
		ups.Message = "Meta上传出错,重试" + strconv.Itoa(index) + "/ 7 原因:" + err.Error()
		upch <- ups
		fmt.Println("Meta上传出错,重试", index, "/ 7", "原因:", err)
	}
	if err != nil {
		ups.Message = "Meta上传错误重试老多次还不行 :("
		ups.Error = errors.New("Meta上传错误" + err.Error())
		ups.Code = -1
		upch <- ups
		return
	}
	ups.Message = "上传完毕！"
	ups.Code = 0
	fmt.Println("上传完毕！")
	ups.Ncd = makeMetaURL(metakey)
	upch <- ups
	return
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

func download(ncd string, thread int, downs DownStatus, downch chan DownStatus) {
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
			downs.Message = "第" + strconv.Itoa(blockMeta.Index) + "块下载出错,重试" + strconv.Itoa(index) + "/ 7" + " 原因：" + err.Error()
			downch <- downs
			fmt.Println("第", blockMeta.Index, "块下载出错,重试", index, "/ 7", "原因：", err)
			blockData, err = imageDownload(blockMeta.URL)
			blockDataSha1 = sha1.Sum(blockData[62:])
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

	downs.Code = 1
	downs.OKNUM = 0

	downch <- downs

	var fmutex sync.RWMutex
	var hmutex sync.RWMutex
	var wg sync.WaitGroup
	blockMetach := make(chan BlockMeta)
	fmt.Println("解析Meta")
	meta, err := getMeta(ncd)
	if err != nil {
		downs.Error = "获取Meta出错。" + err.Error()
		downs.Code = -1
		fmt.Println("获取Meta出错。", err)
		downch <- downs
		return
	}
	downs.BlockNum = len(meta.Block)
	downs.Filename = meta.Filename
	downs.FileSha1 = meta.Sha1
	downs.FileSize = sizeString(meta.Size)
	downch <- downs
	history, err := rHistory(meta.Sha1)
	if err != nil {
		history.FileSha1 = meta.Sha1
	}
	f, err := os.Open(meta.Filename)
	if err != nil {
		f, err = os.Create(meta.Filename)
	}
	if err != nil {
		downs.Error = "创建文件出错。" + err.Error()
		downs.Code = -1
		fmt.Println("创建文件时出错。", err)
		downch <- downs
		return
	}
	// fmt.Println(meta)
	downs.Message = "开始下载!"
	downch <- downs
	fmt.Println("开始下载！")
	for i := 0; i < thread; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for blockMeta := range blockMetach {
				if skip(blockMeta.Index, history.BlockIndex) {
					downs.Message = "第" + strconv.Itoa(blockMeta.Index) + "块已下载，跳过"
					atomic.AddInt32(&downs.OKNUM, 1)
					downch <- downs
					fmt.Println("第", blockMeta.Index, "块已下载，跳过")
					continue
				}
				err := core(blockMeta, &fmutex, f)
				if err != nil {
					downs.Message = "第" + strconv.Itoa(blockMeta.Index) + "块下载失败，已跳过。"
					downch <- downs
					fmt.Println("第", blockMeta.Index, "块下载失败，已跳过。")
					continue
				}
				hmutex.Lock()
				history.BlockIndex = append(history.BlockIndex, blockMeta.Index)
				wHistory(history)
				hmutex.Unlock()
				downs.Message = "第" + strconv.Itoa(blockMeta.Index) + "块下载完成。"
				atomic.AddInt32(&downs.OKNUM, 1)
				downch <- downs
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
		downs.Message = "文件校验失败，下载失败！请重试！"
		downs.Error = "文件校验失败，下载失败！请重试！"
		downs.Code = -1
		downch <- downs
		fmt.Println("文件校验失败，下载失败！请重试！")
		return
	}
	downs.Message = "文件校验通过，下载完成！"
	downs.Code = 0
	downch <- downs
	fmt.Println("文件校验通过，下载完成！")
	return
}

func infoMeta(ncd string) (Info, error) {
	var info Info
	meta, err := getMeta(ncd)
	if err != nil {
		fmt.Println("获取Meta出错。", err)
		info.Error = "获取Meta出错。" + err.Error()
		return info, err
	}
	tm := time.Unix(meta.Time, 0)
	fmt.Println("文件名称： ", meta.Filename)
	fmt.Println("上传时间： ", tm.Format("2006-01-02 15:04:05"))
	fmt.Println("文件大小： ", sizeString(meta.Size))
	fmt.Println("文件哈希：  Sha1", meta.Sha1)
	fmt.Println("文件分块： ", len(meta.Block))
	info.Filename = meta.Filename
	info.FileSha1 = meta.Sha1
	info.FileSize = sizeString(meta.Size)
	info.BlockNum = len(meta.Block)
	info.UploadTime = tm.Format("2006-01-02 15:04:05")
	return info, nil
}
