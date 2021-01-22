package scraper

import (
	"net/http"
	"testing"

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
				Logger:             new(defaultLogger),
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
				Logger:             new(defaultLogger),
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
				Logger:             new(defaultLogger),
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
				Logger:             new(defaultLogger),
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
				Logger:             new(defaultLogger),
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
				Logger:             new(defaultLogger),
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
				Logger:             new(defaultLogger),
			},
		},
		{
			name:    "Custom Logger",
			options: []Option{SetLogger(new(defaultLogger))},
			expected: &Scraper{
				url:                "",
				expectedStatusCode: http.StatusOK,
				targetPrice:        DefaultTargetPrice,
				selector:           DefaultSelector,
				findText:           DefaultFindText,
				maxRetries:         DefaultMaxRetries,
				retrySeconds:       DefaultRetrySeconds,
				Logger:             new(defaultLogger),
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
				Logger:             new(defaultLogger),
			},
		},
		{
			name: "All Options at once",
			options: []Option{
				SetURL("https://myurl.com"),
				SetTargetPrice(10),
				SetLogger(new(defaultLogger)),
				SetExpectedStatusCode(201),
				SetFindText("text"),
				SetSelector(".available"),
				SetMaxRetries(20),
				SetRetrySeconds(30),
				SetLogger(new(defaultLogger)),
			},
			expected: &Scraper{
				url:                "https://myurl.com",
				expectedStatusCode: http.StatusCreated,
				targetPrice:        10,
				selector:           ".available",
				findText:           "text",
				maxRetries:         20,
				retrySeconds:       30,
				Logger:             new(defaultLogger),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			assert.Equal(tt, tc.expected, NewScraper(tc.options...))
		})
	}
}
