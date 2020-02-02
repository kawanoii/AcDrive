package main

import (
	"encoding/json"
	"errors"
)

//Meta 元数据包
type Meta struct {
	Time     int64       `json:"time"`
	Filename string      `json:"filename"`
	Size     int64       `json:"size"`
	Sha1     string      `json:"sha1"`
	Block    []BlockMeta `json:"block"`
}

//BlockMeta 元数据包里的块元
type BlockMeta struct {
	Index  int    `json:"index"`
	URL    string `json:"url"`
	Offset int64  `json:"offset"`
	Sha1   string `json:"sha1"`
	Size   int    `json:"size"`
}

// func (meta Meta) blockOffset(index int) int64 {
// 	var offset int
// 	for _, block := range meta.Block[:index] {
// 		offset += block.Size
// 	}
// 	return int64(offset)
// }

func getMeta(ncd string) (Meta, error) {
	var meta Meta
	metaURL, err := unnCov(ncd)
	if err != nil {
		return meta, errors.New("nCovDrive地址有误。")
	}
	metaDATA, err := imageDownload(metaURL)
	if err != nil {
		return meta, err
	}
	json.Unmarshal(metaDATA[62:], &meta)
	if len(meta.Block) == 0 {
		return meta, errors.New("元数据解析错误！MetaData:" + string(metaDATA[62:]))
	}
	return meta, nil
}
