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
	Phone       string     `json:"from"`
	URL         string     `json:"url"`
	Type        ActionType `json:"type"`
	Price       string     `json:"price,omitempty"`
	SoldOutText string     `json:"sold_out_text,omitempty"`
	Selector    string     `json:"selector,omitempty"`
}

// NewAction returns a new Action object with default values
func NewAction() *Action {
	return &Action{
		SoldOutText: scraper.DefaultSoldOutText,
		Selector:    scraper.DefaultSelector,
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
func (at ActionType) IsValid() error {
	switch at {
	case priceAction, availabilityAction:
		return nil
	}
	return fmt.Errorf("invalid action type: %s", at)
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
	if err := action.Type.IsValid(); err != nil {
		return c.JSON(http.StatusBadRequest, &ResponseMessage{err.Error()})
	}

	var secondParameter string

	switch action.Type {
	case priceAction:
		secondParameter = fmt.Sprintf("%s|%s", action.Selector, action.Price)
	default:
		secondParameter = fmt.Sprintf("%s|%s", action.Selector, action.SoldOutText)
	}

	sugar.Debugw("adding product to database",
		"action", action.Type,
		"phone", action.Phone,
		"url", action.URL,
		"selector", action.Selector,
		"sold_out_text", action.SoldOutText,
		"price", action.Price,
	)

	return db.Save(fmt.Sprintf("%s|%s", action.Phone, action.URL), secondParameter, string(action.Type))
}
