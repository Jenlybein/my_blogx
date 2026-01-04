package io_utils

import (
	"bytes"
	"io"
)

func GetBody(body *io.ReadCloser) ([]byte, error) {
	byteData, err := io.ReadAll(*body)
	if err != nil {
		return nil, err
	}

	// 内容阅后即焚，所以需要恢复请求体内容
	*body = io.NopCloser(bytes.NewBuffer(byteData))

	return byteData, nil
}
