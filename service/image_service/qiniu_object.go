package image_service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"myblogx/global"

	"github.com/qiniu/go-sdk/v7/auth"
	"github.com/qiniu/go-sdk/v7/storage"
)

type qiniuRuntime struct {
	mac        *auth.Credentials
	bucketMgr  *storage.BucketManager
	httpClient *http.Client
}

var (
	qiniuRuntimeMu   sync.Mutex
	qiniuRuntimeKey  string
	qiniuRuntimeInst *qiniuRuntime
)

func getQiniuRuntime() *qiniuRuntime {
	qiniuRuntimeMu.Lock()
	defer qiniuRuntimeMu.Unlock()

	confKey := fmt.Sprintf("%s|%s", global.Config.QiNiu.AccessKey, global.Config.QiNiu.SecretKey)
	if qiniuRuntimeInst == nil || qiniuRuntimeKey != confKey {
		mac := auth.New(global.Config.QiNiu.AccessKey, global.Config.QiNiu.SecretKey)
		cfg := &storage.Config{UseHTTPS: true}
		qiniuRuntimeInst = &qiniuRuntime{
			mac:       mac,
			bucketMgr: storage.NewBucketManager(mac, cfg),
			httpClient: &http.Client{
				Timeout: 10 * time.Second,
			},
		}
		qiniuRuntimeKey = confKey
	}
	return qiniuRuntimeInst
}

func VerifyQiniuCallback(req *http.Request) (bool, error) {
	return getQiniuRuntime().mac.VerifyCallback(req)
}

func StatObject(bucket, key string) (*storage.FileInfo, error) {
	info, err := getQiniuRuntime().bucketMgr.Stat(bucket, key)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

func DeleteObject(bucket, key string) error {
	return getQiniuRuntime().bucketMgr.Delete(bucket, key)
}

func ImageInfoObject(bucket, key string) (*ImageInfoResult, error) {
	_ = bucket

	q := global.Config.QiNiu
	if strings.TrimSpace(q.Uri) == "" {
		return nil, errors.New("七牛下载域名未配置")
	}

	domain := strings.TrimRight(strings.TrimSpace(q.Uri), "/")
	deadline := time.Now().Add(3 * time.Minute).Unix()
	downloadURL := storage.MakePrivateURLv2(getQiniuRuntime().mac, domain, key, deadline)
	if strings.Contains(downloadURL, "?") {
		downloadURL += "&imageInfo"
	} else {
		downloadURL += "?imageInfo"
	}

	resp, err := getQiniuRuntime().httpClient.Get(downloadURL)
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
