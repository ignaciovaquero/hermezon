package twilio

import (
	"fmt"

	"github.com/igvaquero18/hermezon/utils"
	"github.com/sfreiberg/gotwilio"
)

// Notifier is an interface for sending SMS Messages
type Notifier interface {
	SendSMS(from, to, body, statusCallback, applicationSid string, opts ...*gotwilio.Option) (smsResponse *gotwilio.SmsResponse, exception *gotwilio.Exception, err error)
}

// Client is a struct that implicitly implements the Messenger interface
// for sending messages throught Twilio
type Client struct {
	Notifier
	utils.Logger
}

// Option is a function to apply settings to Client structure
type Option func(c *Client) Option

// NewClient returns a new Client. It requires an account SID and an
// authorization token from Twilio.
func NewClient(accountSid, authToken string, opts ...Option) *Client {
	twilio := gotwilio.NewTwilioClient(accountSid, authToken)
	c := &Client{
		Notifier: twilio,
		Logger:   &utils.DefaultLogger{},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// SetLogger Sets the Logger for Scraper
func SetLogger(logger utils.Logger) Option {
	return func(c *Client) Option {
		prev := c.Logger
		c.Logger = logger
		return SetLogger(prev)
	}
}

// SendMessage sends a message throught Twilio API. It is a wrapper for
// implementing the Messenger interface
func (c *Client) SendMessage(title, body, from, dest string) error {
	c.Debug("sending SMS message through Twilio API")
	resp, exception, err := c.SendSMS(from, dest, fmt.Sprintf("%s\n\n%s", title, body), "", "")
	if err != nil {
		return err
	}
	if exception != nil {
		return fmt.Errorf("Twilio API error. code: %d, message: %s", exception.Code, exception.Message)
	}
	c.Debugw("response from Twilio API",
		"date_created", resp.DateCreated,
		"date_sent", resp.DateSent,
		"status", resp.Status,
		"body", resp.Body,
	)
	return nil
}
