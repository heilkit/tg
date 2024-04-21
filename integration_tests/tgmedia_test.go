package integration_tests

import (
	"encoding/json"
	"github.com/heilkit/tg"
	"github.com/heilkit/tg/scheduler"
	"github.com/heilkit/tg/tgmedia"
	"github.com/heilkit/tg/tgvideo"
	"github.com/stretchr/testify/require"
	"log"
	"os"
	"sync"
	"testing"
	"time"
)

var env = load("env.json")

func TestFromDisk(t *testing.T) {
	t.Parallel()

	files := []string{
		"testdata/normal.jpg",
		"testdata/too_big.png",
		"testdata/vid.mp4",
		"testdata/vid.webm",
	}
	bot := env.Bot
	chat := env.Chat

	wg := &sync.WaitGroup{}
	wg.Add(len(files))
	for _, filename := range files {
		filename := filename
		go func() {
			_, err := bot.Send(chat, tgmedia.FromDisk(filename))
			require.NoError(t, err, filename)
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestMetadata(t *testing.T) {
	t.Parallel()

	bot := env.Bot
	chat := env.Chat
	msg, err := bot.Send(chat, tgmedia.FromDisk(
		"testdata/vid.mp4",
		tgvideo.ThumbnailAt("0.1"),
		tgvideo.EmbedMetadata(map[string]string{"key_1": "value_1", "key_2": "value_2"}),
	))
	require.NoError(t, err, "testdata/vid.mp4")

	file, err := bot.DownloadTemp(&msg.Video.File)
	require.NoError(t, err, file.Name())
	defer os.Remove(file.Name())

	meta, err := tgvideo.ExtractMetadata[map[string]string](file.Name())
	require.NoError(t, err)

	require.Equal(t, "value_1", meta["key_1"])
	require.Equal(t, "value_2", meta["key_2"])
}

func TestHandlers(t *testing.T) {
	bot := env.Bot

	go time.AfterFunc(time.Minute*15, func() {
		bot.Stop()
	})

	calls := 0
	bot.HandleAlbum(func(cs tg.Contexts) error {
		calls += 1
		_, e := bot.CopyMessages(cs.Chat(), cs.Messages())
		return e
	}, tg.HandleAlbumByTimeOption)

	bot.Start()
	require.Equal(t, 1, calls)
}

type envT struct {
	Token  string   `json:"token"`
	ChatId int64    `json:"chat"`
	Chat   *tg.Chat `json:"-"`
	Bot    *tg.Bot  `json:"-"`
}

func load(filename string) envT {
	buff, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("file %s reading failed, %v", filename, err)
	}
	var ret envT
	if err := json.Unmarshal(buff, &ret); err != nil {
		log.Fatal(err)
	}

	ret.Chat = &tg.Chat{ID: ret.ChatId}
	ret.Bot, err = tg.NewBot(tg.Settings{
		Token:     ret.Token,
		OnError:   tg.OnErrorLog(),
		Scheduler: scheduler.ExtraConservative(),
	})
	if err != nil {
		log.Fatal(err)

	}

	return ret
}
