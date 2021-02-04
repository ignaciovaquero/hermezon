package main

import (
	"fmt"
	"net/http"

	"github.com/igvaquero18/hermezon/scraper"
	"github.com/labstack/echo/v4"
)

// Action is the action we will perform for tracking
// products
type Action struct {
	From     string     `json:"from"`
	URL      string     `json:"url"`
	Type     ActionType `json:"type"`
	Price    string     `json:"price,omitempty"`
	FindText string     `json:"find_text,omitempty"`
	Selector string     `json:"selector,omitempty"`
}

// NewAction returns a new Action object with default values
func NewAction() *Action {
	return &Action{
		FindText: scraper.DefaultFindText,
		Selector: scraper.DefaultSelector,
	}
}

// ActionType is a wrapper around the string type to define
// what kind of actions we can perform
type ActionType string

const (
	priceAction        = "price"
	availabilityAction = "availability"
)

// IsValid checks whether an action is valid or not
func (at ActionType) IsValid() bool {
	switch at {
	case priceAction, availabilityAction:
		return true
	}
	return false
}

// ResponseMessage is a struct for building responses
type ResponseMessage struct {
	Message string `json:"message"`
}

// postActions will allow us to post a new product for tracking
// its price or availability
func postActions(c echo.Context) error {
	action := NewAction()
	if err := c.Bind(&action); err != nil {
		return c.JSON(http.StatusBadRequest, &ResponseMessage{fmt.Sprintf("invalid action: %s", err.Error())})
	}
	if action.Type.IsValid() {
		return c.JSON(http.StatusBadRequest, &ResponseMessage{"invalid action"})
	}

	var secondParameter string

	switch action.Type {
	case priceAction:
		secondParameter = fmt.Sprintf("%s|%s", action.Selector, action.Price)
	default:
		secondParameter = fmt.Sprintf("%s|%s", action.Selector, action.FindText)
	}

	sugar.Debugw("adding product to database",
		"action", action.Type,
		"from", action.From,
		"url", action.URL,
		"selector", action.Selector,
		"find_text", action.FindText,
		"price", action.Price,
	)

	return db.Save(fmt.Sprintf("%s|%s", action.From, action.URL), secondParameter, string(action.Type))
}
