package email_store

import (
	"sync"
	"time"
)

type EmailVerifyStore struct {
	store        sync.Map
	MaxFailCount int
	timeout      time.Duration
}

type EmailStoreInfo struct {
	Email     string
	Code      string
	FailCount int
	timer     *time.Timer
}

func NewEmailVerifyStore(max int, timeout int) *EmailVerifyStore {
	timeoutDuration := time.Duration(timeout) * time.Minute
	return &EmailVerifyStore{
		MaxFailCount: max,
		timeout:      timeoutDuration,
	}
}

func (s *EmailVerifyStore) Store(id, email, code string) {
	// 停止已有timer
	s.stopTimer(id)

	// 存储新数据
	s.store.Store(id, &EmailStoreInfo{
		Email:     email,
		Code:      code,
		FailCount: 0,
		timer: time.AfterFunc(s.timeout, func() {
			s.store.Delete(id)
		}),
	})
}

func (s *EmailVerifyStore) stopTimer(id string) {
	if info, ok := s.store.Load(id); ok {
		if info, ok := info.(*EmailStoreInfo); ok {
			info.timer.Stop()
		}
	}
}

func (s *EmailVerifyStore) Delete(id string) {
	s.stopTimer(id)
	s.store.Delete(id)
}

func (s *EmailVerifyStore) Verify(id, code string) (info *EmailStoreInfo, ok bool) {
	emailInfo, ok := s.store.Load(id)
	if !ok {
		return
	}

	info, ok = emailInfo.(*EmailStoreInfo)
	if !ok {
		return
	}

	if info.Code != code {
		info.FailCount++
		if info.FailCount >= s.MaxFailCount {
			s.Delete(id)
		}
		return nil, false
	}

	s.Delete(id)

	return info, true
}
