package types

type TelegramUpdate struct {
	UpdateId      int64                 `json:"update_id"`
	Message       TelegramMessage       `json:"message"`
	CallbackQuery TelegramCallbackQuery `json:"callback_query"`
}

type TelegramResponse struct {
	Method             string                      `json:"method"`
	MessageId          int64                       `json:"message_id,omitempty"`
	ChatId             int64                       `json:"chat_id"`
	ParseMode          string                      `json:"parse_mode"`
	Text               string                      `json:"text"`
	ReplyMarkup        any                         `json:"reply_markup,omitempty"`
	LinkPreviewOptions *TelegramLinkPreviewOptions `json:"link_preview_options,omitempty"`
}

type TelegramMessage struct {
	MessageId int64        `json:"message_id"`
	Date      int64        `json:"date"`
	From      TelegramUser `json:"from"`
	Chat      TelegramChat `json:"chat"`
	Text      string       `json:"text"`
}

type TelegramCallbackQuery struct {
	Id      string          `json:"id"`
	From    TelegramUser    `json:"from"`
	Message TelegramMessage `json:"message"`
	Data    string          `json:"data"`
}

type TelegramUser struct {
	Id        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
}

type TelegramChat struct {
	Id        int64  `json:"id"`
	Type      string `json:"type"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type TelegramReplyKeyboardMarkup struct {
	Keyboard       [][]string `json:"keyboard"`
	ResizeKeyboard bool       `json:"resize_keyboard"`
}

var DefaultReplyMarkup = TelegramReplyKeyboardMarkup{
	Keyboard: [][]string{
		{"Watch", "Unwatch"},
		{"Watch List"},
	},
	ResizeKeyboard: true,
}

type TelegramInlineKeyboardMarkup struct {
	InlineKeyboard [][]TelegramInlineKeyboardButton `json:"inline_keyboard"`
}

type TelegramInlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data"`
}

type TelegramLinkPreviewOptions struct {
	IsDisabled bool `json:"is_disabled"`
}
