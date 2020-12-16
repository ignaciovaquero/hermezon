package main

import (
	"fmt"
	"strings"

	"github.com/igvaquero18/hermezon/scraper"
	bolt "go.etcd.io/bbolt"
)

type availability struct{}

// Run checks all the products in the database and tracks its
// availability in the corresponding store
func (a availability) Run() {
	err := db.View(func(tx *bolt.Tx) error {
		sugar.Debug("checking all products for availability")
		b := tx.Bucket([]byte(availabilityAction))
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
			phone := keys[0]
			url := keys[1]
			selector := values[0]
			soldOutText := values[1]
			sugar.Debugw("checking product availability for customer",
				"phone", phone,
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
					sugar.Errorw("error when checking price", "phone", phone, "msg", err.Error())
				}
				if productAvailable {
					sugar.Debugw("Product is available!", "phone", phone, "url", url)
					err = messagingClient.SendMessage(
						"Product is available!",
						fmt.Sprintf("URL: %s", url),
						twilioPhone,
						phone,
					)
					if err != nil {
						sugar.Errorw("error when sending message", "msg", err.Error())
						return
					}
					err = db.Update(func(tx *bolt.Tx) error {
						bucket := tx.Bucket([]byte(availabilityAction))
						if bucket == nil {
							sugar.Warnw("bucket doesn't exist", "bucket", availabilityAction)
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
					sugar.Debugw("deleted key from bucket", "key", k, "bucket", availabilityAction)
					return
				}
				sugar.Debugw("Product is sold out...", "phone", phone, "url", url)
			}()

			return nil
		})
		return nil
	})

	if err != nil {
		sugar.Fatalw("error when reading the database", "msg", err.Error())
	}
}
