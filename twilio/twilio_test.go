package twilio

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	mock_twilio "github.com/igvaquero18/hermezon/twilio/mock_twilio"
	"github.com/igvaquero18/hermezon/utils"
	"github.com/sfreiberg/gotwilio"
	"github.com/stretchr/testify/assert"
	_ "github.com/stretchr/testify/assert"
)

func TestSendMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tw := mock_twilio.NewMockNotifier(ctrl)
	gomock.InOrder(
		tw.EXPECT().SendSMS(
			"+34612345678",
			"+34698765432",
			fmt.Sprintf("%s\n\n%s", "Title", "Body"),
			"",
			"",
		).Return(
			&gotwilio.SmsResponse{
				DateCreated: "2021-02-01",
				DateSent:    "2021-02-01",
				Status:      "Created",
				Body:        "Created",
			},
			nil,
			nil,
		),
		tw.EXPECT().SendSMS(
			"+34612345678",
			"",
			fmt.Sprintf("%s\n\n%s", "Title", "Body"),
			"",
			"",
		).Return(
			&gotwilio.SmsResponse{
				DateCreated: "2021-02-01",
				DateSent:    "2021-02-01",
				Status:      "Created",
				Body:        "Created",
			},
			nil,
			fmt.Errorf("An error ocurred"),
		),
		tw.EXPECT().SendSMS(
			"+34612345678",
			"+3469876543",
			fmt.Sprintf("%s\n\n%s", "Title", "Body"),
			"",
			"",
		).Return(
			&gotwilio.SmsResponse{
				DateCreated: "2021-02-01",
				DateSent:    "2021-02-01",
				Status:      "Created",
				Body:        "Created",
			},
			&gotwilio.Exception{
				Code:    gotwilio.ExceptionCode(1000),
				Message: "Something went wrong",
			},
			nil,
		),
	)

	cl := &Client{
		tw,
		&utils.DefaultLogger{},
	}

	testCases := []struct {
		name   string
		to     string
		err    error
		client *Client
	}{
		{
			name:   "Correct SMS sent",
			to:     "+34698765432",
			err:    nil,
			client: cl,
		},
		{
			name:   "Error when sending SMS",
			to:     "",
			err:    fmt.Errorf("An error ocurred"),
			client: cl,
		},
		{
			name:   "Correct SMS sent",
			to:     "+3469876543",
			err:    fmt.Errorf("Twilio API error"),
			client: cl,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			err := tc.client.SendMessage("Title", "Body", "+34612345678", tc.to)
			if tc.err != nil {
				assert.Error(tt, err)
			} else {
				assert.NoError(tt, err)
			}
		})
	}
}
