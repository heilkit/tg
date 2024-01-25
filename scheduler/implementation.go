package scheduler

import (
	"slices"
	"strconv"
	"sync"
	"time"
)

type scheduler struct {
	globalLimit int
	global      int

	perChatLimit int
	perChat      map[string]int

	sync        *sync.RWMutex
	events      []event
	pollingRate time.Duration
}

var _ Scheduler = &scheduler{}

func (sch *scheduler) SyncFunc(count int, chat string, fn RawFunc) (ret []byte, err error) {
	if sch == nil {
		ret, err = fn()
		return
	}

	ticker := time.NewTicker(sch.pollingRate)
	defer ticker.Stop()
	for now := time.Now(); true; now = <-ticker.C {
		sch.sync.Lock()
		sch.handleEvents(now)

		if !sch.isReadyFor(count, chat) {
			sch.sync.Unlock()
			continue
		}

		ret, err = fn()
		sch.add(count, chat)

		sch.sync.Unlock()
		break
	}

	return
}

func (sch *scheduler) isReadyFor(count int, chat string) bool {
	if sch.globalLimit < sch.global+count {
		return false
	}
	if perChat, contains := sch.perChat[chat]; contains && sch.perChatLimit < perChat+count {
		return false
	}

	return true
}

type event struct {
	time  time.Time
	count int
	chat  string
}

func (sch *scheduler) order() {
	slices.SortFunc(sch.events, func(lhs, rhs event) int {
		return lhs.time.Compare(rhs.time)
	})
}

func (sch *scheduler) add(count int, chat string) {
	now := time.Now()

	sch.global += count
	sch.events = append(sch.events, event{
		time:  now.Add(ApiRequestQuotaTimeout),
		count: count,
		chat:  "",
	})

	if !isPersonal(chat) {
		sch.perChat[chat] += count
		sch.events = append(sch.events, event{
			time:  now.Add(ApiRequestQuotaPerChatTimeout),
			count: count,
			chat:  chat,
		})
	}

	sch.order() // it's not the best implementation
}

func (sch *scheduler) handleEvents(now time.Time) {
	if len(sch.events) == 0 {
		return
	}

	handled := 0
	for _, event := range sch.events {
		if event.time.After(now) {
			break
		}

		handled += 1
		sch.global -= event.count
		if isPersonal(event.chat) {
			continue
		}

		sch.perChat[event.chat] -= event.count
		if sch.perChat[event.chat] <= 0 {
			delete(sch.perChat, event.chat)
		}
	}
	sch.events = sch.events[handled:]
}

func isPersonal(chat string) bool {
	id, err := strconv.ParseInt(chat, 10, 64)
	if err != nil {
		return false
	}
	return id >= 0
}
