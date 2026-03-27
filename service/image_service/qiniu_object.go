package image_service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"myblogx/global"

	"github.com/qiniu/go-sdk/v7/auth"
	"github.com/qiniu/go-sdk/v7/storage"
)

var qiniuBucketManager *storage.BucketManager

func getQiniuBucketManager() *storage.BucketManager {
	if qiniuBucketManager == nil {
		mac := auth.New(global.Config.QiNiu.AccessKey, global.Config.QiNiu.SecretKey)
		cfg := &storage.Config{UseHTTPS: true}
		qiniuBucketManager = storage.NewBucketManager(mac, cfg)
	}
	return qiniuBucketManager
}

func VerifyQiniuCallback(req *http.Request) (bool, error) {
	mac := auth.New(global.Config.QiNiu.AccessKey, global.Config.QiNiu.SecretKey)
	return mac.VerifyCallback(req)
}

func StatObject(bucket, key string) (*storage.FileInfo, error) {
	info, err := getQiniuBucketManager().Stat(bucket, key)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

func DeleteObject(bucket, key string) error {
	return getQiniuBucketManager().Delete(bucket, key)
}

func ImageInfoObject(bucket, key string) (*ImageInfoResult, error) {
	_ = bucket

	q := global.Config.QiNiu
	if strings.TrimSpace(q.Uri) == "" {
		return nil, errors.New("七牛下载域名未配置")
	}

	domain := strings.TrimRight(strings.TrimSpace(q.Uri), "/")
	mac := auth.New(q.AccessKey, q.SecretKey)
	deadline := time.Now().Add(3 * time.Minute).Unix()
	downloadURL := storage.MakePrivateURLv2(mac, domain, key, deadline)
	if strings.Contains(downloadURL, "?") {
		downloadURL += "&imageInfo"
	} else {
		downloadURL += "?imageInfo"
	}

	resp, err := (&http.Client{Timeout: 10 * time.Second}).Get(downloadURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("七牛图片信息读取失败，状态码 %d", resp.StatusCode)
	}

	var result ImageInfoResult
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if result.Format == "" || result.Width <= 0 || result.Height <= 0 {
		return nil, errors.New("七牛图片信息不完整")
	}
	return &result, nil
}

func ObjectURL(key string) string {
	domain := strings.TrimRight(strings.TrimSpace(global.Config.QiNiu.Uri), "/")
	if domain == "" {
		return ""
	}
	if !strings.HasPrefix(domain, "http://") && !strings.HasPrefix(domain, "https://") {
		domain = "https://" + domain
	}
	return storage.MakePublicURLv2(domain, key)
}
