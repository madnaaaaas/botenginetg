package botendinetg

import (
	"fmt"
	"sync"

	tgbotapi "github.com/Syfaro/telegram-bot-api"
)

type user struct {
	sync.Mutex
	state      interface{}
	page       string
	lastMsgIds []int
	engine     *Engine
}

func chatIdFromUpdate(upd *tgbotapi.Update) int64 {
	if upd.CallbackQuery != nil {
		return upd.CallbackQuery.Message.Chat.ID
	}
	if upd.Message != nil {
		return upd.Message.Chat.ID
	}
	return -1
}

func deleteMessages(msgIds []int, chatID int64, bot *tgbotapi.BotAPI) {
	for _, msgId := range msgIds {
		if _, err := bot.DeleteMessage(tgbotapi.NewDeleteMessage(chatID, msgId)); err != nil {
			fmt.Println(err)
		}
	}
}

func sendPhoto(photoData []byte, chatID int64, bot *tgbotapi.BotAPI) (int, error) {
	photoConfig := tgbotapi.NewPhotoUpload(chatID, tgbotapi.FileBytes{Name: "dumb", Bytes: photoData})
	tgMsg, err := bot.Send(photoConfig)
	if err != nil {
		return 0, err
	}
	return tgMsg.MessageID, nil
}

func (usr *user) send(msg Message, upd *tgbotapi.Update, deleteUserMessage bool) {
	chatID := chatIdFromUpdate(upd)
	deleteMessages(usr.lastMsgIds, chatID, usr.engine.bot)
	usr.lastMsgIds = nil
	if msg.Text == "" && msg.PhotoData == nil {
		return
	}
	if msg.PhotoData != nil {
		if tgMsgID, err := sendPhoto(msg.PhotoData, chatID, usr.engine.bot); err != nil {
			fmt.Println(err)
		} else {
			usr.lastMsgIds = append(usr.lastMsgIds, tgMsgID)
		}
	}
	tgMsgConfig := tgbotapi.NewMessage(chatID, msg.Text)
	if msg.Keyboard != nil {
		tgMsgConfig.ReplyMarkup = telegramKeyboard(msg.Keyboard)
	}
	if tgMsg, err := usr.engine.bot.Send(tgMsgConfig); err != nil {
		fmt.Println(err)
	} else {
		usr.lastMsgIds = append(usr.lastMsgIds, tgMsg.MessageID)
	}
	if deleteUserMessage && upd.Message != nil {
		if _, err := usr.engine.bot.DeleteMessage(tgbotapi.NewDeleteMessage(chatID, upd.Message.MessageID)); err != nil {
			fmt.Println(err)
		}
	}
}

func (usr *user) update(upd *tgbotapi.Update) {
	usr.Lock()
	defer usr.Unlock()

	v := Verdict{}
	if usr.engine.GlobalUpdateCallback != nil {
		v = usr.engine.GlobalUpdateCallback(newUpdate(upd, usr.engine.bot), usr.state)
	}
	if v.Page != "" {
		usr.page = v.Page
	} else {
		v = usr.engine.Callbacks[usr.page].Update(newUpdate(upd, usr.engine.bot), usr.state)
		usr.page = v.Page
	}
	msg := usr.engine.Callbacks[usr.page].Show(usr.state)

	usr.send(msg, upd, v.DeleteUserMessage)
}
