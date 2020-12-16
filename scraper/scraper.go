package scraper

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
)

const (
	// DefaultSelector is a default CSS selector for the Availability message
	DefaultSelector = "#availability span"

	// DefaultSoldOutText is the text to compare to check if the item is unavailable.
	DefaultSoldOutText = "no disponible."

	// DefaultMaxRetries is the default number of max retries allowed.
	DefaultMaxRetries int8 = 1

	// DefaultRetrySeconds is the default number of seconds to wait between retries.
	DefaultRetrySeconds int8 = 1

	// DefaultTargetPrice is the default target price.
	DefaultTargetPrice float64 = 0

	userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_6) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0.2 Safari/605.1.15"
)

// Scraper is a struct that contains all the details for performing scraping of a store
// website like Amazon, El Corte Ingles, Scraper, etc.
type Scraper struct {
	url                string
	expectedStatusCode int
	targetPrice        float64
	selector           string
	soldOutText        string
	maxRetries         int8
	retrySeconds       int8
	Logger
}

// Option is a function to apply settings to Scraper structure
type Option func(s *Scraper) Option

// NewScraper returns a new instance of Scraper
func NewScraper(opts ...Option) *Scraper {
	m := &Scraper{
		url:                "",
		expectedStatusCode: http.StatusOK,
		targetPrice:        DefaultTargetPrice,
		selector:           DefaultSelector,
		soldOutText:        DefaultSoldOutText,
		maxRetries:         DefaultMaxRetries,
		retrySeconds:       DefaultRetrySeconds,
		Logger:             new(defaultLogger),
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// SetURL Sets the URL for Scraper
func SetURL(url string) Option {
	return func(s *Scraper) Option {
		prev := s.url
		s.url = url
		return SetURL(prev)
	}
}

// SetLogger Sets the Logger for Scraper
func SetLogger(logger Logger) Option {
	return func(s *Scraper) Option {
		prev := s.Logger
		s.Logger = logger
		return SetLogger(prev)
	}
}

// SetExpectedStatusCode Sets the Logger for Scraper
func SetExpectedStatusCode(expectedStatusCode int) Option {
	return func(s *Scraper) Option {
		prev := s.expectedStatusCode
		s.expectedStatusCode = expectedStatusCode
		return SetExpectedStatusCode(prev)
	}
}

// SetSelector Sets the Logger for Scraper
func SetSelector(selector string) Option {
	return func(s *Scraper) Option {
		prev := s.selector
		s.selector = selector
		return SetSelector(prev)
	}
}

// SetSoldOutText Sets the text to compare
func SetSoldOutText(soldOutText string) Option {
	return func(s *Scraper) Option {
		prev := s.soldOutText
		s.soldOutText = soldOutText
		return SetSoldOutText(prev)
	}
}

// SetMaxRetries Sets the text to compare
func SetMaxRetries(maxRetries int8) Option {
	return func(s *Scraper) Option {
		prev := s.maxRetries
		s.maxRetries = maxRetries
		return SetMaxRetries(prev)
	}
}

// SetRetrySeconds Sets the text to compare
func SetRetrySeconds(retrySeconds int8) Option {
	return func(s *Scraper) Option {
		prev := s.retrySeconds
		if retrySeconds > 0 {
			s.retrySeconds = retrySeconds
		}
		return SetRetrySeconds(prev)
	}
}

// SetTargetPrice Sets the target price for the product
func SetTargetPrice(target float64) Option {
	return func(s *Scraper) Option {
		prev := s.targetPrice
		s.targetPrice = target
		return SetTargetPrice(prev)
	}
}

func (s Scraper) getTextInSelector() (string, error) {
	var retries int8 = 0
	client := new(http.Client)
	req, err := http.NewRequest(http.MethodGet, s.url, nil)
	if err != nil {
		return "", fmt.Errorf("error building the request: %x", err.Error())
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := client.Do(req)
	for (err != nil || resp.StatusCode != s.expectedStatusCode) && retries <= s.maxRetries {
		retries++
		s.Debug("temporary error when accessing store. retrying...")
		seconds, _ := time.ParseDuration(fmt.Sprintf("%ds", s.retrySeconds))
		time.Sleep(seconds)
		resp, err = client.Do(req)
	}

	if err != nil {
		return "", err
	}

	if resp.StatusCode != s.expectedStatusCode {
		return "", fmt.Errorf("response status code: %d, expected: %d", resp.StatusCode, s.expectedStatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	err = resp.Body.Close()
	if err != nil {
		return "", err
	}

	text := doc.Find(s.selector).Text()
	s.Debugw("found text", "text", text, "selector", s.selector)
	return text, nil
}

// IsAvailable checks whether the product is available or not
func (s Scraper) IsAvailable() (bool, error) {
	text, err := s.getTextInSelector()
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(strings.ToLower(s.soldOutText)) != strings.TrimSpace(strings.ToLower(text)), nil
}

// IsPriceBelow returns true if the price is below s.targetPrice
func (s Scraper) IsPriceBelow() (bool, error) {
	text, err := s.getTextInSelector()
	if err != nil {
		return false, err
	}
	if text == "" {
		return false, fmt.Errorf("no price matched")
	}
	reg := regexp.MustCompile(`\d+[\.\,]?\d*`)
	text = strings.ReplaceAll(reg.FindString(text), ",", ".")
	price, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return false, errors.Wrap(err, "error when converting price from string to float")
	}

	return s.targetPrice > price, nil
}
