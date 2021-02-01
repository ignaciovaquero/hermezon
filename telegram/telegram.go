package telegram

import (
	"strconv"

	"github.com/igvaquero18/hermezon/utils"
	"github.com/igvaquero18/telegram-notifier/telegram"
)

// Client wraps around telegram.Client in order to implement Messenger interface
type Client struct {
	Notifier
}

// Notifier is an interface for sending messages to a list of channels
type Notifier interface {
	SendNotification(title, body string, chats []int64) error
}

// NewClient returns a Client. A valid Telegram token must be provided.
// If logger is nil, a logger with basic capabilities will be used instead.
func NewClient(telegramToken string, logger utils.Logger) (*Client, error) {
	if logger == nil {
		logger = &utils.DefaultLogger{}
	}
	t, err := telegram.NewClient(telegramToken, logger)
	if err != nil {
		return nil, err
	}
	return &Client{t}, nil
}

// SendMessage creates a new Telegram Client and sends a message
func (c *Client) SendMessage(title, body, from, dest string) error {
	intDest, err := strconv.ParseInt(dest, 10, 64)
	if err != nil {
		return err
	}
	return c.SendNotification(title, body, []int64{intDest})
}
