package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/igvaquero18/hermezon/scraper"
	bolt "go.etcd.io/bbolt"
)

type price struct{}

// Run checks all the products in the database and tracks its
// price in the corresponding store, looking for price drops.
func (p price) Run() {
	err := db.View(func(tx *bolt.Tx) error {
		sugar.Debug("checking all products for price down")
		b := tx.Bucket([]byte(priceAction))
		if b == nil {
			sugar.Debug("no products matched")
			return nil
		}
		b.ForEach(func(k, v []byte) error {
			keys := strings.Split(string(k), "|")
			values := strings.Split(string(v), "|")
			if len(keys) < 2 || len(values) < 2 {
				return fmt.Errorf("invalid keys and/or values returned from database. keys: %v, values: %v", keys, values)
			}
			reg := regexp.MustCompile(`\d+[\.\,]?\d*`)
			phone := keys[0]
			url := keys[1]
			selector := values[0]
			targetPriceStr := values[1]
			targetPrice, err := strconv.ParseFloat(strings.ReplaceAll(reg.FindString(targetPriceStr), ",", "."), 64)
			if err != nil {
				return fmt.Errorf("invalid target price retrieved from database: %s", err.Error())
			}

			sugar.Debugw("checking product price for customer",
				"phone", phone,
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
					sugar.Errorw("error when checking price", "phone", phone, "msg", err.Error())
				}
				if priceBelow {
					sugar.Debugw("Price is below!", "phone", phone, "url", url, "desired_price", targetPriceStr)
					err = messagingClient.SendMessage(
						"Product is below desired price!",
						fmt.Sprintf("URL: %s\nDesired price: %s", url, targetPriceStr),
						twilioPhone,
						phone,
					)
					if err != nil {
						sugar.Errorw("error when sending message", "msg", err.Error())
						return
					}
					err = db.Update(func(tx *bolt.Tx) error {
						bucket := tx.Bucket([]byte(priceAction))
						if bucket == nil {
							sugar.Warnw("bucket doesn't exist", "bucket", priceAction)
							return nil
						}
						if err = bucket.Delete(k); err != nil {
							sugar.Errorw("error when deleting key from bucket", "key", k, "msg", err.Error())
						}
						return nil
					})
					if err != nil {
						sugar.Fatalw("error when reading the database", "msg", err.Error())
					}
					sugar.Debugw("deleted key from bucket", "key", k, "bucket", priceAction)
					return
				}
				sugar.Debugw("Price is not below...", "phone", phone, "url", url, "desired_price", targetPriceStr)
			}()

			return nil
		})
		return nil
	})
	if err != nil {
		sugar.Fatalw("error when reading the database", "msg", err.Error())
	}
}
