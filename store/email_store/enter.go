package email_store

import (
	"sync"
	"time"
)

type EmailVerifyStore sync.Map

func NewEmailVerifyStore(maxFailCount int, timeout time.Duration) *EmailVerifyStore {
	return (*EmailVerifyStore)(&sync.Map{})
}

type EmailStoreInfo struct {
	Email     string
	Code      string
	FailCount int
}
