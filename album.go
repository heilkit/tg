package tg

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Messages are enhanced list of contexts.
type Messages []Message

func (b *Bot) ForwardMessages(to Recipient, messages []Message, opts ...interface{}) ([]Message, error) {
	if to == nil {
		return nil, ErrBadRecipient
	}
	if len(messages) == 0 {
		return nil, ErrEmptyMessage
	}

	_, chatID := messages[0].MessageSig()
	ids := []string{}
	for _, msg := range messages {
		ids = append(ids, strconv.FormatInt(int64(msg.ID), 10))
	}

	params := map[string]string{
		"chat_id":      to.Recipient(),
		"from_chat_id": strconv.FormatInt(chatID, 10),
		"message_ids":  fmt.Sprintf("[%s]", strings.Join(ids, ",")),
	}

	sendOpts := extractOptions(opts)
	b.embedSendOptions(params, sendOpts)

	data, err := b.Raw("forwardMessages", params)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Result []struct {
			MessageId int `json:"message_id"`
		} `json:"result"`
	}
	if err := extractOk(data); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	ret := []Message{}
	for _, msgID := range resp.Result {
		msg := Message{
			ID:   msgID.MessageId,
			Chat: &Chat{ID: chatID},
		}
		ret = append(ret, msg)
	}
	return ret, nil
}

func (b *Bot) CopyMessages(to Recipient, messages []Message, opts ...interface{}) ([]Message, error) {
	if to == nil {
		return nil, ErrBadRecipient
	}
	if len(messages) == 0 {
		return nil, ErrEmptyMessage
	}

	_, chatID := messages[0].MessageSig()
	ids := []string{}
	for _, msg := range messages {
		ids = append(ids, strconv.FormatInt(int64(msg.ID), 10))
	}

	params := map[string]string{
		"chat_id":      to.Recipient(),
		"from_chat_id": strconv.FormatInt(chatID, 10),
		"message_ids":  fmt.Sprintf("[%s]", strings.Join(ids, ",")),
	}

	sendOpts := extractOptions(opts)
	b.embedSendOptions(params, sendOpts)

	data, err := b.Raw("copyMessages", params)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Result []struct {
			MessageId int `json:"message_id"`
		} `json:"result"`
	}
	if err := extractOk(data); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	ret := []Message{}
	for _, msgID := range resp.Result {
		msg := Message{
			ID:   msgID.MessageId,
			Chat: &Chat{ID: chatID},
		}
		ret = append(ret, msg)
	}
	return ret, nil
}

func (b *Bot) DeleteMessages(messages []Message, opts ...interface{}) error {
	if len(messages) == 0 {
		return ErrEmptyMessage
	}

	_, chatID := messages[0].MessageSig()
	ids := []string{}
	for _, msg := range messages {
		ids = append(ids, strconv.FormatInt(int64(msg.ID), 10))
	}

	params := map[string]string{
		"chat_id":     strconv.FormatInt(chatID, 10),
		"message_ids": fmt.Sprintf("[%s]", strings.Join(ids, ",")),
	}

	sendOpts := extractOptions(opts)
	b.embedSendOptions(params, sendOpts)

	data, err := b.Raw("deleteMessages", params)
	if err != nil {
		return err
	}

	var resp struct {
		Result bool `json:"result"`
	}
	if err := extractOk(data); err != nil {
		return err
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return err
	}
	// could this even happen?
	if !resp.Result {
		return ErrNotFoundToDelete
	}
	return nil
}

// Album lets you group multiple media into a single message.
type Album []Inputtable

func (album Album) WithCaption(text string) Album {
	if len(album) == 0 {
		return album
	}
	album[0].WithCaption(text)
	if len(album) > 1 {
		for _, media := range album[1:] {
			media.WithCaption("")
		}
	}
	return album
}

// Contexts are enhanced list of contexts.
type Contexts []Context

func (contexts Contexts) Chat() *Chat {
	if len(contexts) == 0 {
		return nil
	}
	return contexts[0].Chat()
}

func (contexts Contexts) Album() Album {
	album := Album{}
	for _, ctx := range contexts {
		album = append(album, ctx.Message().InputMedia())
	}
	return album
}

func (contexts Contexts) Messages() []Message {
	messages := []Message{}
	for _, ctx := range contexts {
		messages = append(messages, *ctx.Message())
	}
	return messages
}

func (contexts Contexts) IsAlbum() bool {
	for _, ctx := range contexts {
		if media := ctx.Message().InputMedia(); media == nil {
			return false
		}
	}
	return true
}

// AlbumHandlerFunc is just like HandlerFunc, but for list of contexts.
type AlbumHandlerFunc func(cs Contexts) error

func (f AlbumHandlerFunc) ToHandlerFunc() HandlerFunc {
	return func(c Context) error {
		return f([]Context{c})
	}
}

// HandleAlbum opts -- MiddlewareFunc / endpoints (OnPhoto, OnVideo...) -- default=telebot.OnMedia.
// I.e. bot.HandleAlbum(userHandler, telebot.OnPhoto, telebot.OnVideo, middleware.WhiteList(777)).
// Sadly, there's no way to define both bot.Handle(telebot.OnPhoto,..) and bot.HandleAlbum(telebot.OnPhoto,..).
func (b *Bot) HandleAlbum(handler AlbumHandlerFunc, opts ...interface{}) {
	b.Group().HandleAlbum(handler, opts...)
}

func (g *Group) HandleAlbum(handler AlbumHandlerFunc, opts ...interface{}) {
	endpoints := make([]interface{}, 0)
	middlewares := make([]MiddlewareFunc, 0)
	for _, opt := range opts {
		switch o := opt.(type) {
		case MiddlewareFunc:
			middlewares = append(middlewares, o)
		default:
			endpoints = append(endpoints, o)
		}
	}
	if len(endpoints) == 0 {
		endpoints = append(endpoints, OnMedia)
	}

	delay := time.Second / 2
	var albumHandler handleManager
	if g.b.synchronous {
		albumHandler = newSyncedManager(g.b, handler, delay)
	} else {
		albumHandler = newUnsyncedManager(delay, handler)
	}

	for _, endpoint := range endpoints {
		g.Handle(endpoint, func(ctx Context) error {
			return albumHandler.add(ctx)
		}, middlewares...)
	}
}

type handleManager interface {
	add(ctx Context) error
}

var _ handleManager = &syncedManager{}
var _ handleManager = &unsyncedManager{}

type syncedManager struct {
	bot     *Bot
	fn      AlbumHandlerFunc
	delay   time.Duration
	current string
	ctx     []Context
	sync    *sync.Mutex
}

func newSyncedManager(bot *Bot, fn AlbumHandlerFunc, delay time.Duration) *syncedManager {
	return &syncedManager{
		bot:     bot,
		fn:      fn,
		delay:   delay,
		current: "",
		ctx:     nil,
		sync:    &sync.Mutex{},
	}
}

func (manager *syncedManager) delayHandling(id string) {
	go func() {
		time.Sleep(manager.delay)

		manager.sync.Lock()
		defer manager.sync.Unlock()

		if len(manager.ctx) == 0 {
			return
		}

		defer func() {
			if r := recover(); r != nil {
				manager.bot.onError(fmt.Errorf("syncedManager.delayHandling(id) panicked: %v", r), manager.ctx[0])
			}
		}()

		if id != manager.current {
			return
		}

		if err := manager.fn(manager.ctx); err != nil {
			manager.bot.onError(err, manager.ctx[0])
		}

		manager.current = ""
		manager.ctx = nil
	}()
}

func (manager *syncedManager) add(ctx Context) (err error) {
	manager.sync.Lock()
	defer manager.sync.Unlock()

	msg := ctx.Message()
	id := mediaGroupToId(msg)
	if manager.current == id {
		manager.ctx = append(manager.ctx, ctx)
		return
	}

	if len(manager.ctx) != 0 {
		err = manager.fn(manager.ctx)
	}
	manager.current = id
	manager.ctx = []Context{ctx}

	manager.delayHandling(id)

	return
}

type handleSchedulerUnit struct {
	delays int
	ctx    []Context
}

type unsyncedManager struct {
	handler         AlbumHandlerFunc
	delay           time.Duration
	unscheduled     map[string]handleSchedulerUnit
	unscheduledSync *sync.Mutex
}

func newUnsyncedManager(timeout time.Duration, handler AlbumHandlerFunc) *unsyncedManager {
	return &unsyncedManager{
		handler:         handler,
		delay:           timeout,
		unscheduled:     map[string]handleSchedulerUnit{},
		unscheduledSync: &sync.Mutex{},
	}
}

func (handleScheduler *unsyncedManager) add(ctx Context) error {
	handleScheduler.unscheduledSync.Lock()
	defer handleScheduler.unscheduledSync.Unlock()

	id := mediaGroupToId(ctx.Message())
	if unit, ok := handleScheduler.unscheduled[id]; ok {
		unit.ctx = append(unit.ctx, ctx)
		unit.delays += 1
		handleScheduler.unscheduled[id] = unit
		go time.AfterFunc(handleScheduler.delay, func() { handleScheduler.handle(id) })
		return nil
	}

	handleScheduler.unscheduled[id] = handleSchedulerUnit{
		delays: 1,
		ctx:    []Context{ctx},
	}
	go time.AfterFunc(handleScheduler.delay, func() { handleScheduler.handle(id) })

	return nil
}

func (handleScheduler *unsyncedManager) handle(id string) {
	handleScheduler.unscheduledSync.Lock()
	defer handleScheduler.unscheduledSync.Unlock()

	unit, ok := handleScheduler.unscheduled[id]
	if !ok {
		return
	}
	unit.delays -= 1
	handleScheduler.unscheduled[id] = unit

	if unit.delays == 0 {
		defer func() {
			delete(handleScheduler.unscheduled, id)
			if r := recover(); r != nil {
				ctx := unit.ctx[0]
				ctx.Bot().OnError(fmt.Errorf("album handling paniced: %v", r), ctx)
			}
		}()

		contexts := unit.ctx
		sort.Slice(contexts, func(i, j int) bool { return contexts[i].Message().ID < contexts[j].Message().ID })

		if err := handleScheduler.handler(unit.ctx); err != nil {
			ctx := unit.ctx[0]
			ctx.Bot().OnError(err, ctx)
		}
	}
}

func singleMessage(msg *Message) bool {
	return msg.AlbumID == ""
}

func mediaGroupToId(msg *Message) string {
	if !singleMessage(msg) {
		return msg.AlbumID
	}
	return fmt.Sprintf("%d_%d", msg.Chat.ID, msg.ID)
}
