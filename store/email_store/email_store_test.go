package email_store_test

import (
	"myblogx/store/email_store"
	"testing"
	"time"
)

func TestEmailVerifyStoreSuccess(t *testing.T) {
	s := email_store.NewEmailVerifyStore(3, 1)
	s.Store("id1", "a@example.com", "123456")

	info, ok := s.Verify("id1", "123456")
	if !ok {
		t.Fatal("验证码应校验成功")
	}
	if info.Email != "a@example.com" {
		t.Fatalf("邮箱不一致: %s", info.Email)
	}

	if _, ok = s.Verify("id1", "123456"); ok {
		t.Fatal("成功校验后应被删除")
	}
}

func TestEmailVerifyStoreFailCount(t *testing.T) {
	s := email_store.NewEmailVerifyStore(2, 1)
	s.Store("id2", "b@example.com", "654321")

	if _, ok := s.Verify("id2", "000000"); ok {
		t.Fatal("错误验证码不应通过")
	}
	if _, ok := s.Verify("id2", "000000"); ok {
		t.Fatal("错误验证码不应通过")
	}
	if _, ok := s.Verify("id2", "654321"); ok {
		t.Fatal("超出失败次数后应删除")
	}
}

func TestEmailVerifyStoreTimeoutAndDelete(t *testing.T) {
	// timeout=0 分钟，立刻过期
	s := email_store.NewEmailVerifyStore(3, 0)
	s.Store("id3", "c@example.com", "111111")
	time.Sleep(10 * time.Millisecond)
	if _, ok := s.Verify("id3", "111111"); ok {
		t.Fatal("超时后不应通过")
	}

	s.Store("id4", "d@example.com", "222222")
	s.Delete("id4")
	if _, ok := s.Verify("id4", "222222"); ok {
		t.Fatal("Delete 后不应通过")
	}
}
