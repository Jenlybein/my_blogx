package qiniu_service

import (
	"context"
	"myblogx/global"
	"time"

	"github.com/qiniu/go-sdk/v7/storagev2/credentials"
	"github.com/qiniu/go-sdk/v7/storagev2/uptoken"
)

func GenUpToken(key string) (upToken string, err error) {
	q := global.Config.QiNiu

	mac := credentials.NewCredentials(q.AccessKey, q.SecretKey)

	putPolicy, err := uptoken.NewPutPolicyWithKey(
		q.Bucket,
		key,
		time.Now().Add(time.Duration(q.Expiry)*time.Second),
	)
	if err != nil {
		return "", err
	}

	upToken, err = uptoken.NewSigner(putPolicy, mac).GetUpToken(context.Background())
	if err != nil {
		return "", err
	}

	return upToken, nil
}
