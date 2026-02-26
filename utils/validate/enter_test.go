package validate_test

import (
	"io"
	"myblogx/utils/validate"
	"testing"
)

func TestValidateErrorHelpers(t *testing.T) {
	msg := validate.ValidateErr(io.EOF)
	if msg == "" {
		t.Fatal("ValidateErr 返回空字符串")
	}

	data, msg2 := validate.ValidateError(io.EOF)
	if data != nil {
		t.Fatalf("普通错误 data 应为 nil: %+v", data)
	}
	if msg2 == "" {
		t.Fatal("ValidateError msg 不应为空")
	}
}
