package integration_tests

import (
	"encoding/json"
	"github.com/heilkit/tg"
	"github.com/heilkit/tg/scheduler"
	"github.com/heilkit/tg/tgmedia"
	"github.com/stretchr/testify/require"
	"log"
	"os"
	"sync"
	"testing"
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
