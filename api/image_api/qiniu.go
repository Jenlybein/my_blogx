package image_api

import (
	"fmt"
	"myblogx/common/res"
	"myblogx/global"
	"myblogx/service/qiniu_service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type QiNiuGenUpTokenResponse struct {
	Token  string `json:"token"`
	Key    string `json:"key"`
	Region string `json:"region"`
	Url    string `json:"url"`
	Size   int    `json:"size"`
}

func (I *ImageApi) GenUpToken(c *gin.Context) {
	q := global.Config.QiNiu
	if !q.Enable {
		res.FailWithMsg("未启用七牛云配置", c)
		return
	}

	filename := uuid.New().String()
	// TODO：后缀获取
	key := fmt.Sprintf("%s/%s.png", q.Prefix, filename)
	url := fmt.Sprintf("%s/%s", q.Uri, key)

	token, err := qiniu_service.GenUpToken(key)
	if err != nil {
		res.FailWithMsg(err.Error(), c)
		return
	}

	res.OkWithData(&QiNiuGenUpTokenResponse{
		Token:  token,
		Key:    key,
		Region: q.Region,
		Url:    url,
		Size:   q.Size,
	}, c)
}
