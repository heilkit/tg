package scheduler

import (
	"sync"
	"time"
)

const (
	// ApiRequestQuota per second
	ApiRequestQuota        = 30
	ApiRequestQuotaTimeout = time.Second

	// ApiRequestQuotaPerChat per minute
	ApiRequestQuotaPerChat        = 20
	ApiRequestQuotaPerChatTimeout = time.Minute

	DefaultPollingRate = time.Millisecond * 10
)

type RawFunc func() ([]byte, error)

type Scheduler interface {
	SyncFunc(count int, chat string, fn RawFunc) ([]byte, error)
}

// Default Telegram API limits, 20/minute -- per group chat quota, 30/second -- global quota.
func Default() Scheduler {
	return Custom(ApiRequestQuota, ApiRequestQuotaPerChat, DefaultPollingRate)
}

// Conservative gives you a headroom of 20% compared to Default, just in case something goes wrong.
func Conservative() Scheduler {
	return Custom(ApiRequestQuota*4/5, ApiRequestQuotaPerChat*4/5, DefaultPollingRate*10)
}

// ExtraConservative gives you a headroom of 50% compared to Default, just in case something goes wrong.
// Required, if you are planning on uploading tons of files in a chat continuously.
func ExtraConservative() Scheduler {
	return Custom(ApiRequestQuota/2, ApiRequestQuotaPerChat/2, DefaultPollingRate*100)
}

// Custom Telegram API limits, global -- per second, perChat -- per minute.
func Custom(global int, perChat int, pollingRate time.Duration) Scheduler {
	return &scheduler{
		globalLimit:  global,
		global:       0,
		perChatLimit: perChat,
		perChat:      map[string]int{},
		sync:         &sync.RWMutex{},
		events:       []event{},
		pollingRate:  pollingRate,
	}
}

// Nil scheduler does nothing, performing all functions ASAP.
func Nil() Scheduler {
	return &nilScheduler{}
}

func (sch *nilScheduler) SyncFunc(count int, chat string, fn RawFunc) ([]byte, error) {
	return fn()
}

type nilScheduler struct{}

var _ Scheduler = &nilScheduler{}
