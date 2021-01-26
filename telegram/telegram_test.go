package telegram

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	testCases := []struct {
		name   string
		token  string
		log    Logger
		err    error
		notNil bool
	}{
		{
			name:   "string token and nil logger",
			token:  "token",
			log:    nil,
			notNil: true,
			err:    nil,
		},
		{
			name:   "empty token and nil logger",
			token:  "",
			log:    nil,
			notNil: true,
			err:    nil,
		},
		{
			name:   "string token and default logger",
			token:  "token",
			log:    new(defaultLogger),
			notNil: true,
			err:    nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			actual, err := NewClient(tc.token, tc.log)
			if tc.err != nil {
				assert.NoError(tt, err)
			} else {
				assert.Error(tt, err)
			}
			if tc.notNil {
				assert.NotNil(tt, actual)
			}
		})
	}
}
