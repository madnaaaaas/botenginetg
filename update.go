package botendinetg

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	tgbotapi "github.com/Syfaro/telegram-bot-api"
)

const maxPhotoSize = 19*(1 << 20)

var (
	ErrorPhotoNotFound = fmt.Errorf("photo not presented in update")
	ErrorPhotoOverMaxSize = fmt.Errorf("photo over max size (%d Mb)", maxPhotoSize >> 20)
)

type Update struct {
	bot *tgbotapi.BotAPI
	tgUpdate *tgbotapi.Update
}

func newUpdate(tgUpdate *tgbotapi.Update, bot *tgbotapi.BotAPI) *Update {
	return &Update{bot, tgUpdate}
}

func (u *Update) GetMessageText() string {
	if u.tgUpdate.Message != nil {
		return u.tgUpdate.Message.Text
	}

	return ""
}

func (u *Update) GetCallbackQueryText() string {
	if u.tgUpdate.CallbackQuery != nil {
		return u.tgUpdate.CallbackQuery.Data
	}

	return ""
}

func (u *Update) GetPhotoData() ([]byte, error) {
	if u.tgUpdate.Message == nil {
		return nil, ErrorPhotoNotFound
	}
	photoSize := u.tgUpdate.Message.Photo
	if photoSize == nil || len(*photoSize) == 0 {
		return nil, ErrorPhotoNotFound
	}
	maxId := -1
	for i := 0; i < len(*photoSize); i++ {
		photo := (*photoSize)[i]
		if photo.FileSize > maxPhotoSize {
			continue
		}
		if maxId < 0 {
			maxId = i
			continue
		}
		max := (*photoSize)[maxId]
		if photo.FileSize > max.FileSize {
			maxId = i
		}
	}
	if maxId < 0 {
		return nil, ErrorPhotoOverMaxSize
	}
	photo := (*photoSize)[maxId]
	path, err := u.bot.GetFileDirectURL(photo.FileID)
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err = resp.Body.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	buf := &bytes.Buffer{}
	if _, err = io.Copy(buf, resp.Body); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
