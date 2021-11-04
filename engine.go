package botenginetg

import tgbotapi "github.com/Syfaro/telegram-bot-api"

type Button struct {
	Text, Data string
}

type InlineKeyboard [][]Button

func telegramKeyboard(ik InlineKeyboard) tgbotapi.InlineKeyboardMarkup {
	keyboard := tgbotapi.InlineKeyboardMarkup{}
	var row []tgbotapi.InlineKeyboardButton
	for _, ikr := range ik {
		for _, ike := range ikr {
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(ike.Text, ike.Data))
		}
		keyboard.InlineKeyboard, row = append(keyboard.InlineKeyboard, row), nil
	}
	return keyboard
}

type Message struct {
	Text      string
	PhotoData []byte
	Keyboard  InlineKeyboard
}

type Verdict struct {
	Page string
	DeleteUserMessage bool
}

func NewVerdictWithDeletion (page string) Verdict {
	return Verdict{page, true}
}

func NewVerdictWithoutDeletion (page string) Verdict {
	return Verdict{page, false}
}

type Callback struct {
	Show   func(state interface{}) Message
	Update func(update *Update, state interface{}) Verdict
}

type Engine struct {
	Callbacks            map[string]Callback
	GlobalUpdateCallback func(*Update, interface{}) Verdict
	NewState             func(string) interface{}

	users map[string]*user
	bot   *tgbotapi.BotAPI
}

func NewBotEngineTg(token string, newState func(string) interface{}, callbacks map[string]Callback,
	globalUpdate func(*Update, interface{}) Verdict) (*Engine, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	return &Engine{
		users:                make(map[string]*user),
		NewState:             newState,
		bot:                  bot,
		Callbacks:            callbacks,
		GlobalUpdateCallback: globalUpdate,
	}, nil
}

func userName(update *tgbotapi.Update) string {
	if update.Message != nil {
		return update.Message.From.UserName
	}
	if update.CallbackQuery != nil {
		return update.CallbackQuery.From.UserName
	}
	return ""
}

func (engine *Engine) Run() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := engine.bot.GetUpdatesChan(u)
	if err != nil {
		return
	}

	for update := range updates {
		updPtr := &update
		username := userName(updPtr)
		usr, ok := engine.users[username]
		if !ok {
			usr = &user{
				state:      engine.NewState(username),
				page:       "start",
				lastMsgIds: nil,
				engine:     engine,
			}
			engine.users[username] = usr
		}
		go usr.update(updPtr)
	}
}
