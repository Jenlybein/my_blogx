package validate_test

import (
	"io"
	"myblogx/utils/validate"
	"testing"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
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

type validateReq struct {
	Name string `json:"name" label:"名称" binding:"required,min=2"`
	Age  int    `json:"age" label:"年龄" binding:"gte=1"`
}

func TestValidateErrWithValidationErrors(t *testing.T) {
	v := binding.Validator.Engine().(*validator.Validate)
	err := v.Struct(validateReq{Name: "a", Age: 0})
	if err == nil {
		t.Fatal("期望产生校验错误")
	}

	msg := validate.ValidateErr(err)
	if msg == "" {
		t.Fatal("ValidateErr 返回为空")
	}

	data, m := validate.ValidateError(err)
	if len(data) == 0 || m == "" {
		t.Fatalf("ValidateError 返回异常: data=%v msg=%s", data, m)
	}
	if _, ok := data["name"]; !ok {
		t.Fatalf("应包含 name 错误, data=%v", data)
	}
	if _, ok := data["age"]; !ok {
		t.Fatalf("应包含 age 错误, data=%v", data)
	}
}
