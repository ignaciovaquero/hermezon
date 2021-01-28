package telegram

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	mock_telegram "github.com/igvaquero18/hermezon/telegram/mock_telegram"
	"github.com/stretchr/testify/assert"
)

func TestSendMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tel := mock_telegram.NewMockNotifier(ctrl)

	gomock.InOrder(
		tel.EXPECT().SendNotification("Hello", "World", []int64{20}).Return(nil),
		tel.EXPECT().SendNotification("", "World", []int64{20}).Return(nil),
		tel.EXPECT().SendNotification("", "", []int64{20}).Return(nil),
	)

	cl := &Client{tel}

	testCases := []struct {
		name, title, body, from, dest string
		client                        *Client
		err                           error
	}{
		{
			name:   "normal message to proper destination",
			title:  "Hello",
			body:   "World",
			from:   "",
			dest:   "20",
			client: cl,
		},
		{
			name:   "message without title to proper destination",
			title:  "",
			body:   "World",
			from:   "",
			dest:   "20",
			client: cl,
		},
		{
			name:   "totally empty message to proper destination",
			title:  "",
			body:   "",
			from:   "",
			dest:   "20",
			client: cl,
		},
		{
			name:   "normal message to invalid destination",
			title:  "Hello",
			body:   "World",
			from:   "",
			dest:   "2A0",
			client: cl,
			err:    errors.New("An error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			err := tc.client.SendMessage(tc.title, tc.body, tc.from, tc.dest)
			if tc.err != nil {
				assert.Error(tt, err)
			} else {
				assert.NoError(tt, err)
			}
		})
	}
}
