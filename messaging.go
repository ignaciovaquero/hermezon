package main

// Messenger is an interface for sending messages to a channel
type Messenger interface {
	SendMessage(title, body, from, dest string) error
}
