package qiniu_service

import (
	"context"
	"fmt"
	"io"
	"myblogx/global"
	"myblogx/utils/file"
	"myblogx/utils/hash"

	"github.com/google/uuid"
	"github.com/qiniu/go-sdk/v7/storagev2/credentials"
	"github.com/qiniu/go-sdk/v7/storagev2/http_client"
	"github.com/qiniu/go-sdk/v7/storagev2/uploader"
)

func SendFileLocal(localFile string) (err error) {
	mac := credentials.NewCredentials(global.Config.QiNiu.AccessKey, global.Config.QiNiu.SecretKey)

	hash, err := hash.FileMd5(localFile)
	if err != nil {
		return err
	}

	suffix := file.GetImageSuffix(localFile)
	filename := fmt.Sprintf("%s.%s", hash, suffix)

	key := fmt.Sprintf("%s/%s", global.Config.QiNiu.Prefix, filename)
	uploadManager := uploader.NewUploadManager(&uploader.UploadManagerOptions{
		Options: http_client.Options{
			Credentials: mac,
		},
	})

	err = uploadManager.UploadFile(context.Background(), localFile, &uploader.ObjectOptions{
		BucketName: global.Config.QiNiu.Bucket,
		ObjectName: &key,
		FileName:   filename,
	}, nil)

	return err
}

func SendFileReader(reader io.Reader) (err error) {
	mac := credentials.NewCredentials(global.Config.QiNiu.AccessKey, global.Config.QiNiu.SecretKey)

	uid := uuid.New().String()

	filename := fmt.Sprintf("%s.%s", uid, "png")

	key := fmt.Sprintf("%s/%s", global.Config.QiNiu.Prefix, filename)
	uploadManager := uploader.NewUploadManager(&uploader.UploadManagerOptions{
		Options: http_client.Options{
			Credentials: mac,
		},
	})

	err = uploadManager.UploadReader(context.Background(), reader, &uploader.ObjectOptions{
		BucketName: global.Config.QiNiu.Bucket,
		ObjectName: &key,
		FileName:   "",
	}, nil)

	return err
}
