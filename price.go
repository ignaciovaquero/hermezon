package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/igvaquero18/hermezon/scraper"
)

type price struct{}

// Run checks all the products in the database and tracks its
// price in the corresponding store, looking for price drops.
func (p price) Run() {
	sugar.Debug("checking all products for price drop")
	results, err := db.GetAll(priceAction)
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
		reg := regexp.MustCompile(`\d+[\.\,]?\d*`)
		channel := keys[0]
		url := keys[1]
		selector := values[0]
		targetPriceStr := values[1]
		targetPrice, err := strconv.ParseFloat(strings.ReplaceAll(reg.FindString(targetPriceStr), ",", "."), 64)
		if err != nil {
			sugar.Fatalw("invalid target price retrieved from database: %s", err.Error())
		}

		sugar.Debugw("checking product price for customer",
			"channel", channel,
			"url", url,
			"selector", selector,
			"target_price", targetPriceStr,
		)

		// Build the scraper
		scr := scraper.NewScraper(
			scraper.SetExpectedStatusCode(expectedStatusCode),
			scraper.SetLogger(sugar),
			scraper.SetMaxRetries(maxRetries),
			scraper.SetRetrySeconds(retrySeconds),
			scraper.SetSelector(selector),
			scraper.SetTargetPrice(targetPrice),
			scraper.SetURL(url),
		)

		go func() {
			priceBelow, err := scr.IsPriceBelow()
			if err != nil {
				sugar.Errorw("error when checking price", "channel", channel, "url", url, "msg", err.Error())
			}
			if priceBelow {
				sugar.Debugw("Price is below!", "channel", channel, "url", url, "desired_price", targetPriceStr)
				err = messagingClient.SendMessage(
					"Product is below desired price!",
					fmt.Sprintf("URL: %s\nDesired price: %s", url, targetPriceStr),
					twilioPhone,
					channel,
				)
				if err != nil {
					sugar.Errorw("error when sending message", "msg", err.Error())
					return
				}
				err = db.Delete(databaseKey, priceAction)
				if err != nil {
					sugar.Fatalw("error when reading the database", "msg", err.Error())
				}
				sugar.Debugw("deleted key from bucket", "key", databaseKey, "bucket", priceAction)
				return
			}
			sugar.Debugw("Price is not below...", "channel", channel, "url", url, "desired_price", targetPriceStr)
		}()
	}
}
