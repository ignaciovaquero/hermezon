package main

import (
	"fmt"
	"strings"

	"github.com/igvaquero18/hermezon/scraper"
)

type availability struct{}

// Run checks all the products in the database and tracks its
// availability in the corresponding store
func (a availability) Run() {
	sugar.Debug("checking all products for availability")
	results, err := db.GetAll(availabilityAction)
	if err != nil {
		sugar.Fatalw("error when reading the database", "msg", err.Error())
	}
	for k, v := range results {
		databaseKey := k
		keys := strings.Split(databaseKey, "|")
		values := strings.Split(v, "|")
		if len(keys) < 2 || len(values) < 2 {
			sugar.Fatalw("invalid keys and/or values returned from database.", "keys", keys, "values", values)
		}
		channel := keys[0]
		url := keys[1]
		selector := values[0]
		soldOutText := values[1]
		sugar.Debugw("checking product availability for customer",
			"channel", channel,
			"url", url,
			"selector", selector,
			"sold_out_text", soldOutText,
		)

		// Build the scraper
		scr := scraper.NewScraper(
			scraper.SetExpectedStatusCode(expectedStatusCode),
			scraper.SetLogger(sugar),
			scraper.SetMaxRetries(maxRetries),
			scraper.SetRetrySeconds(retrySeconds),
			scraper.SetSelector(selector),
			scraper.SetSoldOutText(soldOutText),
			scraper.SetURL(url),
		)

		go func() {
			productAvailable, err := scr.IsAvailable()
			if err != nil {
				sugar.Errorw("error when checking availability", "channel", channel, "url", url, "msg", err.Error())
			}
			if productAvailable {
				sugar.Debugw("Product is available!", "channel", channel, "url", url)
				err = messagingClient.SendMessage(
					"Product is available!",
					fmt.Sprintf("URL: %s", url),
					twilioPhone,
					channel,
				)
				if err != nil {
					sugar.Errorw("error when sending message", "msg", err.Error())
					return
				}
				err = db.Delete(databaseKey, availabilityAction)
				if err != nil {
					sugar.Fatalw("error when reading the database", "msg", err.Error())
				}
				sugar.Debugw("deleted key from bucket", "key", databaseKey, "bucket", availabilityAction)
				return
			}
			sugar.Debugw("Product is sold out...", "channel", channel, "url", url)
		}()
	}
}
