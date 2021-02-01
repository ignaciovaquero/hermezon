package scraper

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/igvaquero18/hermezon/utils"
	"github.com/stretchr/testify/assert"
)

func TestNewScraper(t *testing.T) {
	testCases := []struct {
		name     string
		options  []Option
		expected *Scraper
	}{
		{
			name:    "No options",
			options: []Option{},
			expected: &Scraper{
				url:                "",
				expectedStatusCode: http.StatusOK,
				targetPrice:        DefaultTargetPrice,
				selector:           DefaultSelector,
				findText:           DefaultFindText,
				maxRetries:         DefaultMaxRetries,
				retrySeconds:       DefaultRetrySeconds,
				Logger:             &utils.DefaultLogger{},
				client:             new(http.Client),
			},
		},
		{
			name:    "Custom price",
			options: []Option{SetTargetPrice(10)},
			expected: &Scraper{
				url:                "",
				expectedStatusCode: http.StatusOK,
				targetPrice:        10.0,
				selector:           DefaultSelector,
				findText:           DefaultFindText,
				maxRetries:         DefaultMaxRetries,
				retrySeconds:       DefaultRetrySeconds,
				Logger:             &utils.DefaultLogger{},
				client:             new(http.Client),
			},
		},
		{
			name:    "Custom Status Code",
			options: []Option{SetExpectedStatusCode(201)},
			expected: &Scraper{
				url:                "",
				expectedStatusCode: http.StatusCreated,
				targetPrice:        DefaultTargetPrice,
				selector:           DefaultSelector,
				findText:           DefaultFindText,
				maxRetries:         DefaultMaxRetries,
				retrySeconds:       DefaultRetrySeconds,
				Logger:             &utils.DefaultLogger{},
				client:             new(http.Client),
			},
		},
		{
			name:    "Custom Selector",
			options: []Option{SetSelector(".available")},
			expected: &Scraper{
				url:                "",
				expectedStatusCode: http.StatusOK,
				targetPrice:        DefaultTargetPrice,
				selector:           ".available",
				findText:           DefaultFindText,
				maxRetries:         DefaultMaxRetries,
				retrySeconds:       DefaultRetrySeconds,
				Logger:             &utils.DefaultLogger{},
				client:             new(http.Client),
			},
		},
		{
			name:    "Custom Text to Find",
			options: []Option{SetFindText("text")},
			expected: &Scraper{
				url:                "",
				expectedStatusCode: http.StatusOK,
				targetPrice:        DefaultTargetPrice,
				selector:           DefaultSelector,
				findText:           "text",
				maxRetries:         DefaultMaxRetries,
				retrySeconds:       DefaultRetrySeconds,
				Logger:             &utils.DefaultLogger{},
				client:             new(http.Client),
			},
		},
		{
			name:    "Custom Max Retries",
			options: []Option{SetMaxRetries(20)},
			expected: &Scraper{
				url:                "",
				expectedStatusCode: http.StatusOK,
				targetPrice:        DefaultTargetPrice,
				selector:           DefaultSelector,
				findText:           DefaultFindText,
				maxRetries:         20,
				retrySeconds:       DefaultRetrySeconds,
				Logger:             &utils.DefaultLogger{},
				client:             new(http.Client),
			},
		},
		{
			name:    "Custom Retry Seconds",
			options: []Option{SetRetrySeconds(30)},
			expected: &Scraper{
				url:                "",
				expectedStatusCode: http.StatusOK,
				targetPrice:        DefaultTargetPrice,
				selector:           DefaultSelector,
				findText:           DefaultFindText,
				maxRetries:         DefaultMaxRetries,
				retrySeconds:       30,
				Logger:             &utils.DefaultLogger{},
				client:             new(http.Client),
			},
		},
		{
			name:    "Custom Logger",
			options: []Option{SetLogger(&utils.DefaultLogger{})},
			expected: &Scraper{
				url:                "",
				expectedStatusCode: http.StatusOK,
				targetPrice:        DefaultTargetPrice,
				selector:           DefaultSelector,
				findText:           DefaultFindText,
				maxRetries:         DefaultMaxRetries,
				retrySeconds:       DefaultRetrySeconds,
				Logger:             &utils.DefaultLogger{},
				client:             new(http.Client),
			},
		},
		{
			name:    "Custom URL",
			options: []Option{SetURL("https://myurl.com")},
			expected: &Scraper{
				url:                "https://myurl.com",
				expectedStatusCode: http.StatusOK,
				targetPrice:        DefaultTargetPrice,
				selector:           DefaultSelector,
				findText:           DefaultFindText,
				maxRetries:         DefaultMaxRetries,
				retrySeconds:       DefaultRetrySeconds,
				Logger:             &utils.DefaultLogger{},
				client:             new(http.Client),
			},
		},
		{
			name: "All Options at once",
			options: []Option{
				SetURL("https://myurl.com"),
				SetTargetPrice(10),
				SetLogger(&utils.DefaultLogger{}),
				SetExpectedStatusCode(201),
				SetFindText("text"),
				SetSelector(".available"),
				SetMaxRetries(20),
				SetRetrySeconds(30),
				SetLogger(&utils.DefaultLogger{}),
			},
			expected: &Scraper{
				url:                "https://myurl.com",
				expectedStatusCode: http.StatusCreated,
				targetPrice:        10,
				selector:           ".available",
				findText:           "text",
				maxRetries:         20,
				retrySeconds:       30,
				Logger:             &utils.DefaultLogger{},
				client:             new(http.Client),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			assert.Equal(tt, tc.expected, NewScraper(tc.options...))
		})
	}
}

type RoundTripFunc func(req *http.Request) (*http.Response, error)

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

func TestGetTextInSelector(t *testing.T) {
	testCases := []struct {
		name         string
		scr          *Scraper
		err          error
		expectedText string
	}{
		{
			name: "no errors, expected status code and empty text",
			scr: &Scraper{
				url:                "https://test.com",
				expectedStatusCode: http.StatusOK,
				targetPrice:        DefaultTargetPrice,
				selector:           "#test",
				findText:           DefaultFindText,
				maxRetries:         DefaultMaxRetries,
				retrySeconds:       DefaultRetrySeconds,
				Logger:             &utils.DefaultLogger{},
				client: NewTestClient(func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewBufferString(`<div id="test"></div>`)),
						Header:     make(http.Header),
					}, nil
				}),
			},
			err:          nil,
			expectedText: "",
		},
		{
			name: "no errors, expected status code and some text in id",
			scr: &Scraper{
				url:                "https://test.com",
				expectedStatusCode: http.StatusOK,
				targetPrice:        DefaultTargetPrice,
				selector:           "#test",
				findText:           DefaultFindText,
				maxRetries:         DefaultMaxRetries,
				retrySeconds:       DefaultRetrySeconds,
				Logger:             &utils.DefaultLogger{},
				client: NewTestClient(func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewBufferString(`<div id="test">something</div>`)),
						Header:     make(http.Header),
					}, nil
				}),
			},
			err:          nil,
			expectedText: "something",
		},
		{
			name: "no errors, expected status code and some text in class",
			scr: &Scraper{
				url:                "https://test.com",
				expectedStatusCode: http.StatusOK,
				targetPrice:        DefaultTargetPrice,
				selector:           ".test",
				findText:           DefaultFindText,
				maxRetries:         DefaultMaxRetries,
				retrySeconds:       DefaultRetrySeconds,
				Logger:             &utils.DefaultLogger{},
				client: NewTestClient(func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewBufferString(`<div class="test">something</div>`)),
						Header:     make(http.Header),
					}, nil
				}),
			},
			err:          nil,
			expectedText: "something",
		},
		{
			name: "no errors, expected status code and some text in span inside div",
			scr: &Scraper{
				url:                "https://test.com",
				expectedStatusCode: http.StatusOK,
				targetPrice:        DefaultTargetPrice,
				selector:           ".test span",
				findText:           DefaultFindText,
				maxRetries:         DefaultMaxRetries,
				retrySeconds:       DefaultRetrySeconds,
				Logger:             &utils.DefaultLogger{},
				client: NewTestClient(func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewBufferString(`<div class="test"><span class="another">something</span></div>`)),
						Header:     make(http.Header),
					}, nil
				}),
			},
			err:          nil,
			expectedText: "something",
		},
		{
			name: "no errors, not expected status code",
			scr: &Scraper{
				url:                "https://test.com",
				expectedStatusCode: http.StatusCreated,
				targetPrice:        DefaultTargetPrice,
				selector:           ".test",
				findText:           DefaultFindText,
				maxRetries:         1,
				retrySeconds:       0,
				Logger:             &utils.DefaultLogger{},
				client: NewTestClient(func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewBufferString(`<div class="test">something</div>`)),
						Header:     make(http.Header),
					}, nil
				}),
			},
			err:          fmt.Errorf("response status code: %d, expected: %d", http.StatusOK, http.StatusCreated),
			expectedText: "something",
		},
		{
			name: "response errors, expected status code",
			scr: &Scraper{
				url:                "https://test.com",
				expectedStatusCode: http.StatusOK,
				targetPrice:        DefaultTargetPrice,
				selector:           ".test",
				findText:           DefaultFindText,
				maxRetries:         1,
				retrySeconds:       0,
				Logger:             &utils.DefaultLogger{},
				client: NewTestClient(func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewBufferString(`<div class="test">something</div>`)),
						Header:     make(http.Header),
					}, fmt.Errorf("some error")
				}),
			},
			err:          fmt.Errorf("some error"),
			expectedText: "something",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			text, err := tc.scr.getTextInSelector()
			if tc.err != nil {
				assert.Error(tt, err)
			} else {
				assert.NoError(tt, err)
				assert.Equal(tt, tc.expectedText, text)
			}
		})
	}
}

func TestIsAvailable(t *testing.T) {
	testCases := []struct {
		name     string
		scr      *Scraper
		err      error
		expected bool
	}{
		{
			name: "no errors and true",
			scr: &Scraper{
				url:                "https://test.com",
				expectedStatusCode: http.StatusOK,
				targetPrice:        DefaultTargetPrice,
				selector:           ".test",
				findText:           "something",
				maxRetries:         DefaultMaxRetries,
				retrySeconds:       DefaultRetrySeconds,
				Logger:             &utils.DefaultLogger{},
				client: NewTestClient(func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewBufferString(`<div class="test">something</div>`)),
						Header:     make(http.Header),
					}, nil
				}),
			},
			err:      nil,
			expected: true,
		},
		{
			name: "no errors and false",
			scr: &Scraper{
				url:                "https://test.com",
				expectedStatusCode: http.StatusOK,
				targetPrice:        DefaultTargetPrice,
				selector:           ".test",
				findText:           "something",
				maxRetries:         DefaultMaxRetries,
				retrySeconds:       DefaultRetrySeconds,
				Logger:             &utils.DefaultLogger{},
				client: NewTestClient(func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewBufferString(`<div class="test">bad</div>`)),
						Header:     make(http.Header),
					}, nil
				}),
			},
			err:      nil,
			expected: false,
		},
		{
			name: "errors",
			scr: &Scraper{
				url:                "https://test.com",
				expectedStatusCode: http.StatusCreated,
				targetPrice:        DefaultTargetPrice,
				selector:           ".test",
				findText:           "something",
				maxRetries:         DefaultMaxRetries,
				retrySeconds:       DefaultRetrySeconds,
				Logger:             &utils.DefaultLogger{},
				client: NewTestClient(func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewBufferString(`<div class="test">something</div>`)),
						Header:     make(http.Header),
					}, nil
				}),
			},
			err:      fmt.Errorf("some error"),
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			actual, err := tc.scr.IsAvailable()
			if tc.err != nil {
				assert.Error(tt, err)
			} else {
				assert.NoError(tt, err)
				assert.Equal(tt, tc.expected, actual)
			}
		})
	}

}

func TestIsPriceBelow(t *testing.T) {
	testCases := []struct {
		name     string
		scr      *Scraper
		err      error
		expected bool
	}{
		{
			name: "no errors and price in euros with commas is below",
			scr: &Scraper{
				url:                "https://test.com",
				expectedStatusCode: http.StatusOK,
				targetPrice:        100.5,
				selector:           ".test",
				findText:           "something",
				maxRetries:         DefaultMaxRetries,
				retrySeconds:       DefaultRetrySeconds,
				Logger:             &utils.DefaultLogger{},
				client: NewTestClient(func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewBufferString(`<div class="test">99,5€</div>`)),
						Header:     make(http.Header),
					}, nil
				}),
			},
			expected: true,
			err:      nil,
		},
		{
			name: "no errors and price in euros with dots is below",
			scr: &Scraper{
				url:                "https://test.com",
				expectedStatusCode: http.StatusOK,
				targetPrice:        100.5,
				selector:           ".test",
				findText:           "something",
				maxRetries:         DefaultMaxRetries,
				retrySeconds:       DefaultRetrySeconds,
				Logger:             &utils.DefaultLogger{},
				client: NewTestClient(func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewBufferString(`<div class="test">99.5  €</div>`)),
						Header:     make(http.Header),
					}, nil
				}),
			},
			expected: true,
			err:      nil,
		},
		{
			name: "no errors and price in euros with commas is above",
			scr: &Scraper{
				url:                "https://test.com",
				expectedStatusCode: http.StatusOK,
				targetPrice:        100.5,
				selector:           ".test",
				findText:           "something",
				maxRetries:         DefaultMaxRetries,
				retrySeconds:       DefaultRetrySeconds,
				Logger:             &utils.DefaultLogger{},
				client: NewTestClient(func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewBufferString(`<div class="test">995,10 €</div>`)),
						Header:     make(http.Header),
					}, nil
				}),
			},
			expected: false,
			err:      nil,
		},
		{
			name: "no errors and price in euros with dots is above",
			scr: &Scraper{
				url:                "https://test.com",
				expectedStatusCode: http.StatusOK,
				targetPrice:        100.5,
				selector:           ".test",
				findText:           "something",
				maxRetries:         DefaultMaxRetries,
				retrySeconds:       DefaultRetrySeconds,
				Logger:             &utils.DefaultLogger{},
				client: NewTestClient(func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewBufferString(`<div class="test">995.10   €  </div>`)),
						Header:     make(http.Header),
					}, nil
				}),
			},
			expected: false,
			err:      nil,
		},
		{
			name: "no errors and price is not found",
			scr: &Scraper{
				url:                "https://test.com",
				expectedStatusCode: http.StatusOK,
				targetPrice:        100.5,
				selector:           ".test",
				findText:           "something",
				maxRetries:         DefaultMaxRetries,
				retrySeconds:       DefaultRetrySeconds,
				Logger:             &utils.DefaultLogger{},
				client: NewTestClient(func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewBufferString(`<div class="text">99,5€</div>`)),
						Header:     make(http.Header),
					}, nil
				}),
			},
			expected: false,
			err:      fmt.Errorf("no price matched"),
		},
		{
			name: "no errors and price in dollars with commas is below",
			scr: &Scraper{
				url:                "https://test.com",
				expectedStatusCode: http.StatusOK,
				targetPrice:        100.5,
				selector:           ".test",
				findText:           "something",
				maxRetries:         DefaultMaxRetries,
				retrySeconds:       DefaultRetrySeconds,
				Logger:             &utils.DefaultLogger{},
				client: NewTestClient(func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewBufferString(`<div class="test">$99,5</div>`)),
						Header:     make(http.Header),
					}, nil
				}),
			},
			expected: true,
			err:      nil,
		},
		{
			name: "no errors and price in pounds with dots is above",
			scr: &Scraper{
				url:                "https://test.com",
				expectedStatusCode: http.StatusOK,
				targetPrice:        100.5,
				selector:           ".test",
				findText:           "something",
				maxRetries:         DefaultMaxRetries,
				retrySeconds:       DefaultRetrySeconds,
				Logger:             &utils.DefaultLogger{},
				client: NewTestClient(func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewBufferString(`<div class="test">£  995.10</div>`)),
						Header:     make(http.Header),
					}, nil
				}),
			},
			expected: false,
			err:      nil,
		},
		{
			name: "no errors and trailing comma",
			scr: &Scraper{
				url:                "https://test.com",
				expectedStatusCode: http.StatusOK,
				targetPrice:        100.5,
				selector:           ".test",
				findText:           "something",
				maxRetries:         DefaultMaxRetries,
				retrySeconds:       DefaultRetrySeconds,
				Logger:             &utils.DefaultLogger{},
				client: NewTestClient(func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewBufferString(`<div class="test">£  99,</div>`)),
						Header:     make(http.Header),
					}, nil
				}),
			},
			expected: true,
			err:      nil,
		},
		{
			name: "no errors and leading comma",
			scr: &Scraper{
				url:                "https://test.com",
				expectedStatusCode: http.StatusOK,
				targetPrice:        100.5,
				selector:           ".test",
				findText:           "something",
				maxRetries:         DefaultMaxRetries,
				retrySeconds:       DefaultRetrySeconds,
				Logger:             &utils.DefaultLogger{},
				client: NewTestClient(func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewBufferString(`<div class="test">£  ,99</div>`)),
						Header:     make(http.Header),
					}, nil
				}),
			},
			expected: true,
			err:      nil,
		},
		{
			name: "errors when getting text in selector",
			scr: &Scraper{
				url:                "https://test.com",
				expectedStatusCode: http.StatusOK,
				targetPrice:        100.5,
				selector:           ".test",
				findText:           "something",
				maxRetries:         DefaultMaxRetries,
				retrySeconds:       DefaultRetrySeconds,
				Logger:             &utils.DefaultLogger{},
				client: NewTestClient(func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewBufferString(`<div class="test">99,5 €</div>`)),
						Header:     make(http.Header),
					}, fmt.Errorf("an error")
				}),
			},
			expected: false,
			err:      fmt.Errorf("an error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			actual, err := tc.scr.IsPriceBelow()
			if tc.err != nil {
				assert.Error(tt, err)
			} else {
				assert.NoError(tt, err)
				assert.Equal(tt, tc.expected, actual)
			}
		})
	}
}
