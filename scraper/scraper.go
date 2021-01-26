package scraper

import (
	"fmt"
	"io/ioutil"
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
	DefaultSelector = "#availability"

	// DefaultFindText is the text to compare to check if the item is available.
	DefaultFindText = "en stock."

	// DefaultMaxRetries is the default number of max retries allowed.
	DefaultMaxRetries int8 = 1

	// DefaultRetrySeconds is the default number of seconds to wait between retries.
	DefaultRetrySeconds int8 = 1

	// DefaultTargetPrice is the default target price.
	DefaultTargetPrice float64 = 0

	userAgent = "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:84.0) Gecko/20100101 Firefox/84.0"
)

// Scraper is a struct that contains all the details for performing scraping of a store
// website like Amazon, El Corte Ingles, Scraper, etc.
type Scraper struct {
	url                string
	expectedStatusCode int
	targetPrice        float64
	selector           string
	findText           string
	maxRetries         int8
	retrySeconds       int8
	client             *http.Client
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
		findText:           DefaultFindText,
		maxRetries:         DefaultMaxRetries,
		retrySeconds:       DefaultRetrySeconds,
		Logger:             new(defaultLogger),
		client:             new(http.Client),
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

// SetFindText Sets the text to compare
func SetFindText(findText string) Option {
	return func(s *Scraper) Option {
		prev := s.findText
		s.findText = findText
		return SetFindText(prev)
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
	req, err := http.NewRequest(http.MethodGet, s.url, nil)
	if err != nil {
		return "", fmt.Errorf("error building the request: %x", err.Error())
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := s.client.Do(req)
	for (err != nil || resp.StatusCode != s.expectedStatusCode) && retries < s.maxRetries {
		retries++
		s.Debug("temporary error when accessing store. retrying...")
		seconds, _ := time.ParseDuration(fmt.Sprintf("%ds", s.retrySeconds))
		time.Sleep(seconds)
		resp, err = s.client.Do(req)
	}

	if err != nil {
		return "", err
	}

	if resp.StatusCode != s.expectedStatusCode {
		body, _ := ioutil.ReadAll(resp.Body)
		s.Debugw(
			"error when accessing store.",
			"url", s.url,
			"selector", s.selector,
			"find_text", s.findText,
			"response_status_code", resp.StatusCode,
			"expected_status_code", s.expectedStatusCode,
			"response_body", string(body),
		)
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
	s.Debugw("found text", "url", s.url, "text", text, "selector", s.selector)
	return text, nil
}

// IsAvailable checks whether the product is available or not
func (s Scraper) IsAvailable() (bool, error) {
	text, err := s.getTextInSelector()
	if err != nil {
		return false, err
	}
	return strings.Contains(strings.TrimSpace(strings.ToLower(text)), strings.TrimSpace(strings.ToLower(s.findText))), nil
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
